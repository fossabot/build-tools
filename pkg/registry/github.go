package registry

import (
	"context"
	"docker.io/go-docker/api/types"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"gitlab.com/sparetimecoders/build-tools/pkg/docker"
	"io"
)

type Github struct {
	dockerRegistry
	Username     string `yaml:"username" env:"GITHUB_USERNAME"`
	Password     string `yaml:"password" env:"GITHUB_PASSWORD"`
	Organisation string `yaml:"organisation" env:"GITHUB_ORGANISATION"`
	Repository   string `yaml:"repository" env:"GITHUB_REPOSITORY"`
}

var _ Registry = &Github{}

func (r Github) Name() string {
	return "Github"
}

func (r Github) Configured() bool {
	return len(r.Repository) > 0
}

func (r Github) Login(client docker.Client, out io.Writer) error {
	if ok, err := client.RegistryLogin(context.Background(), types.AuthConfig{Username: r.Username, Password: r.Password, ServerAddress: "docker.pkg.github.com"}); err == nil {
		_, _ = fmt.Fprintln(out, ok.Status)
		return nil
	} else {
		return err
	}
}

func (r Github) GetAuthInfo() string {
	auth := types.AuthConfig{Username: r.Username, Password: r.Password, ServerAddress: "docker.pkg.github.com"}
	authBytes, _ := json.Marshal(auth)
	return base64.URLEncoding.EncodeToString(authBytes)
}

func (r Github) RegistryUrl() string {
	return fmt.Sprintf("docker.pkg.github.com/%s/%s", r.Organisation, r.Repository)
}

func (r *Github) Create(repository string) error {
	return nil
}
