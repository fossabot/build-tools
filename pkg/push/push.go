package push

import (
	docker2 "docker.io/go-docker"
	"flag"
	"fmt"
	"github.com/liamg/tml"
	"gitlab.com/sparetimecoders/build-tools/pkg/config"
	"gitlab.com/sparetimecoders/build-tools/pkg/docker"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

func Push(dir string, out, eout io.Writer, args ...string) int {
	var dockerfile string
	const (
		defaultDockerfile = "Dockerfile"
		usage             = "name of the Dockerfile to use"
	)
	set := flag.NewFlagSet("push", flag.ExitOnError)
	set.StringVar(&dockerfile, "file", defaultDockerfile, usage)
	set.StringVar(&dockerfile, "f", defaultDockerfile, usage+" (shorthand)")
	_ = set.Parse(args)

	client, err := docker2.NewEnvClient()
	if err != nil {
		_, _ = fmt.Fprintln(eout, tml.Sprintf("<red>%s</red>", err.Error()))
		return -1
	}
	cfg, err := config.Load(dir, out)
	if err != nil {
		_, _ = fmt.Fprintln(eout, tml.Sprintf("<red>%s</red>", err.Error()))
		return -2
	}
	return doPush(client, cfg, dir, dockerfile, out, eout)
}

func doPush(client docker.Client, cfg *config.Config, dir, dockerfile string, out, eout io.Writer) int {
	currentCI := cfg.CurrentCI()
	currentRegistry := cfg.CurrentRegistry()

	if err := currentRegistry.Login(client, out); err != nil {
		_, _ = fmt.Fprintln(eout, tml.Sprintf("<red>%s</red>", err.Error()))
		return -3
	}

	auth := currentRegistry.GetAuthInfo()

	if err := currentRegistry.Create(currentCI.BuildName()); err != nil {
		_, _ = fmt.Fprintln(eout, tml.Sprintf("<red>%s</red>", err.Error()))
		return -4
	}

	content, err := ioutil.ReadFile(filepath.Join(dir, dockerfile))
	if err != nil {
		_, _ = fmt.Fprintln(eout, tml.Sprintf("<red>%s</red>", err.Error()))
		return -5
	}
	stages := docker.FindStages(string(content))

	var tags []string
	for _, stage := range stages {
		tags = append(tags, docker.Tag(currentRegistry.RegistryUrl(), currentCI.BuildName(), stage))
	}

	if dockerTag := os.Getenv("DOCKER_TAG"); len(dockerTag) > 0 {
		tag := docker.Tag(currentRegistry.RegistryUrl(), currentCI.BuildName(), dockerTag)
		tags = append(tags, tag)
		_, _ = fmt.Fprintf(out, "overriding docker tags with value from env DOCKER_TAG %s\n", dockerTag)

	} else {
		tags = append(tags,
			docker.Tag(currentRegistry.RegistryUrl(), currentCI.BuildName(), currentCI.Commit()),
			docker.Tag(currentRegistry.RegistryUrl(), currentCI.BuildName(), currentCI.BranchReplaceSlash()),
		)
		if currentCI.Branch() == "master" {
			tags = append(tags, docker.Tag(currentRegistry.RegistryUrl(), currentCI.BuildName(), "latest"))
		}
	}
	for _, tag := range tags {
		_, _ = fmt.Fprintln(out, tml.Sprintf("Pushing tag '<green>%s</green>'", tag))
		if err := currentRegistry.PushImage(client, auth, tag, out, eout); err != nil {
			_, _ = fmt.Fprintln(eout, tml.Sprintf("<red>%s</red>", err.Error()))
			return -6
		}
	}
	return 0
}
