package scaffold

import (
	"errors"
	"fmt"
	"github.com/liamg/tml"
	"gitlab.com/sparetimecoders/build-tools/pkg/config/scaffold/ci"
	"gitlab.com/sparetimecoders/build-tools/pkg/config/scaffold/vcs"
	"gitlab.com/sparetimecoders/build-tools/pkg/file"
	"gitlab.com/sparetimecoders/build-tools/pkg/stack"
	"gitlab.com/sparetimecoders/build-tools/pkg/templating"
	"io"
	"net/url"
	"path/filepath"
	"strings"
)

type Config struct {
	VCS          *VCSConfig `yaml:"vcs"`
	CI           *CIConfig  `yaml:"ci"`
	RegistryUrl  string     `yaml:"registry" env:"REGISTRY"`
	Organisation string     `yaml:"organisation"`
	currentCI    ci.CI
	currentVCS   vcs.VCS
}

type VCSConfig struct {
	Selected string      `yaml:"selected" env:"VCS"`
	Github   *vcs.Github `yaml:"github"`
	Gitlab   *vcs.Gitlab `yaml:"gitlab"`
}

type CIConfig struct {
	Selected  string        `yaml:"selected" env:"CI"`
	Buildkite *ci.Buildkite `yaml:"buildkite"`
	Gitlab    *ci.Gitlab    `yaml:"gitlab"`
}

func (c *Config) Configure() error {
	c.currentVCS.Configure()
	return c.currentCI.Configure()
}

func (c *Config) ValidateConfig() error {
	switch c.VCS.Selected {
	case "github":
		c.currentVCS = c.VCS.Github
	case "gitlab":
		c.currentVCS = c.VCS.Gitlab
	default:
		return errors.New("no VCS configured")
	}
	switch c.CI.Selected {
	case "buildkite":
		c.currentCI = c.CI.Buildkite
	case "gitlab":
		c.currentCI = c.CI.Gitlab
	default:
		return errors.New("no CI configured")
	}

	if err := c.currentVCS.ValidateConfig(); err != nil {
		return err
	}
	return c.currentCI.ValidateConfig()
}

func (c *Config) Validate(name string) error {
	if err := c.currentVCS.Validate(name); err != nil {
		return err
	}
	return c.currentCI.Validate(name)
}

func (c *Config) Scaffold(dir, name string, stack stack.Stack, out io.Writer) int {
	_, _ = fmt.Fprint(out, tml.Sprintf("<lightblue>Creating new service </lightblue><white><bold>'%s'</bold></white> <lightblue>using stack </lightblue><white><bold>'%s'</bold></white>\n", name, stack.Name()))
	_, _ = fmt.Fprint(out, tml.Sprintf("<lightblue>Creating repository at </lightblue><white><bold>'%s'</bold></white>\n", c.currentVCS.Name()))
	repository, err := c.currentVCS.Scaffold(name)
	if err != nil {
		_, _ = fmt.Fprintln(out, tml.Sprintf("<red>%s</red>", err.Error()))
		return -5
	}
	_, _ = fmt.Fprint(out, tml.Sprintf("<green>Created repository </green><white><bold>'%s'</bold></white>\n", repository.SSHURL))
	if err := c.currentVCS.Clone(dir, name, repository.SSHURL, out); err != nil {
		_, _ = fmt.Fprintln(out, tml.Sprintf("<red>%s</red>", err.Error()))
		return -6
	}
	projectDir := filepath.Join(dir, name)
	_, _ = fmt.Fprint(out, tml.Sprintf("<lightblue>Creating build pipeline for </lightblue><white><bold>'%s'</bold></white>\n", name))
	badges, err := c.currentCI.Badges(name)
	if err != nil {
		_, _ = fmt.Fprintln(out, tml.Sprintf("<red>%s</red>", err.Error()))
		return -7
	}
	parsedUrl, err := url.Parse(repository.HTTPURL)
	if err != nil {
		_, _ = fmt.Fprintln(out, tml.Sprintf("<red>%s</red>", err.Error()))
		return -8
	}
	data := templating.TemplateData{
		ProjectName:    name,
		Badges:         badges,
		Organisation:   c.Organisation,
		RegistryUrl:    c.RegistryUrl,
		RepositoryUrl:  repository.SSHURL,
		RepositoryHost: parsedUrl.Host,
		RepositoryPath: strings.Replace(parsedUrl.Path, ".git", "", 1),
	}
	webhook, err := c.currentCI.Scaffold(dir, data)
	if err != nil {
		_, _ = fmt.Fprintln(out, tml.Sprintf("<red>%s</red>", err.Error()))
		return -9
	}
	if err := addWebhook(name, webhook, c.currentVCS); err != nil {
		_, _ = fmt.Fprintln(out, tml.Sprintf("<red>%s</red>", err.Error()))
		return -10
	}
	if err := createDotfiles(projectDir); err != nil {
		_, _ = fmt.Fprintln(out, tml.Sprintf("<red>%s</red>", err.Error()))
		return -11
	}
	if err := createReadme(projectDir, data); err != nil {
		_, _ = fmt.Fprintln(out, tml.Sprintf("<red>%s</red>", err.Error()))
		return -12
	}
	if err := createDeployment(projectDir, data); err != nil {
		_, _ = fmt.Fprintln(out, tml.Sprintf("<red>%s</red>", err.Error()))
		return -13
	}
	if err := stack.Scaffold(projectDir, data); err != nil {
		_, _ = fmt.Fprintln(out, tml.Sprintf("<red>%s</red>", err.Error()))
		return -14
	}
	return 0
}

func InitEmptyConfig() *Config {
	return &Config{
		VCS: &VCSConfig{
			Selected: "",
			Github:   &vcs.Github{},
			Gitlab:   &vcs.Gitlab{},
		},
		CI: &CIConfig{
			Selected:  "",
			Buildkite: &ci.Buildkite{},
			Gitlab:    &ci.Gitlab{},
		},
	}
}

func addWebhook(name string, url *string, vcs vcs.VCS) error {
	if url != nil {
		return vcs.Webhook(name, *url)
	}
	return nil
}

func createDotfiles(dir string) error {
	if err := file.Write(dir, ".gitignore", ""); err != nil {
		return err
	}
	editorconfig := `
root = true

[*]
end_of_line = lf
insert_final_newline = true
charset = utf-8
trim_trailing_whitespace = true
`
	if err := file.Write(dir, ".editorconfig", editorconfig); err != nil {
		return err
	}
	dockerignore := `
.git
.editorconfig
Dockerfile
README.md
`
	if err := file.Write(dir, ".dockerignore", dockerignore); err != nil {
		return err
	}
	return nil
}

func createReadme(dir string, data templating.TemplateData) error {
	content := `
| README.md
# {{.ProjectName}}
{{range .Badges}}[![{{.Title}}]({{.ImageUrl}})]({{.LinkUrl}}){{end}}
`
	return file.WriteTemplated(dir, "README.md", content, data)
}

func createDeployment(dir string, data templating.TemplateData) error {
	return file.WriteTemplated(dir, filepath.Join("deployment_files", "deploy.yaml"), deployment, data)
}

var deployment = `
apiVersion: apps/v1
kind: Deployment
metadata:
 labels:
   app: {{ .ProjectName }}
 name: {{ .ProjectName }}
 annotations:
   kubernetes.io/change-cause: "${TIMESTAMP} Deployed commit id: ${COMMIT}"
spec:
 replicas: 2
 selector:
   matchLabels:
     app: {{ .ProjectName }}
 strategy:
   rollingUpdate:
     maxSurge: 1
     maxUnavailable: 1
   type: RollingUpdate
 template:
   metadata:
     labels:
       app: {{ .ProjectName }}
   spec:
     affinity:
       podAntiAffinity:
         preferredDuringSchedulingIgnoredDuringExecution:
         - weight: 100
           podAffinityTerm:
             labelSelector:
               matchExpressions:
               - key: "app"
                 operator: In
                 values:
                 - {{ .ProjectName }}
             topologyKey: kubernetes.io/hostname
     containers:
     - name: {{ .ProjectName }}
       readinessProbe:
         httpGet:
           path: /
           port: 80
         initialDelaySeconds: 5
         periodSeconds: 5
         timeoutSeconds: 5
       imagePullPolicy: Always
       image: {{ .RegistryUrl }}/{{ .ProjectName }}:${COMMIT}
       ports:
       - containerPort: 80
     restartPolicy: Always
---

apiVersion: v1
kind: Service
metadata:
 name: {{ .ProjectName }}
spec:
 ports:
 - port: 80
   protocol: TCP
   targetPort: 80
 selector:
   app: {{ .ProjectName }}
 type: ClusterIP
`