package config

import (
	"context"
	"docker.io/go-docker/api/types"
	"encoding/base64"
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	awsecr "github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/ecr/ecriface"
	"gitlab.com/sparetimecoders/build-tools/pkg/docker"
	"log"
	"strings"
)

type ECRRegistry struct {
	dockerRegistry
	Url      string `yaml:"url" env:"ECR_URL"`
	Region   string `yaml:"region" env:"ECR_REGION"`
	username string
	password string
	svc      ecriface.ECRAPI
}

var _ Registry = &ECRRegistry{}

func (r *ECRRegistry) configured() bool {
	if len(r.Url) > 0 {
		sess, err := session.NewSession(&aws.Config{Region: &r.Region})
		if err != nil {
			return false
		}
		r.svc = awsecr.New(sess)
		return true
	}
	return false
}

func (r *ECRRegistry) Login(client docker.Client) error {
	input := &awsecr.GetAuthorizationTokenInput{}

	result, err := r.svc.GetAuthorizationToken(input)
	if err != nil {
		return err
	}

	decoded, err := base64.StdEncoding.DecodeString(*result.AuthorizationData[0].AuthorizationToken)
	if err != nil {
		return err
	}
	parts := strings.Split(string(decoded), ":")
	r.username = parts[0]
	r.password = parts[1]

	if ok, err := client.RegistryLogin(context.Background(), types.AuthConfig{Username: r.username, Password: r.password, ServerAddress: r.Url}); err == nil {
		log.Println(ok.Status)
		return nil
	} else {
		return err
	}
}

func (r *ECRRegistry) GetAuthInfo() string {
	auth := types.AuthConfig{Username: r.username, Password: r.password}
	authBytes, _ := json.Marshal(auth)
	return base64.URLEncoding.EncodeToString(authBytes)
}

func (r ECRRegistry) RegistryUrl() string {
	return r.Url
}

func (r ECRRegistry) Create(repository string) error {
	if _, err := r.svc.DescribeRepositories(&awsecr.DescribeRepositoriesInput{RepositoryNames: []*string{&repository}}); err != nil {
		input := &awsecr.CreateRepositoryInput{
			RepositoryName: aws.String(repository),
		}

		if _, err := r.svc.CreateRepository(input); err != nil {
			return err
		} else {
			policyText := `{"rules":[{"rulePriority":10,"description":"Only keep 20 images","selection":{"tagStatus":"untagged","countType":"imageCountMoreThan","countNumber":20},"action":{"type":"expire"}}]}`
			if _, err := r.svc.PutLifecyclePolicy(&awsecr.PutLifecyclePolicyInput{LifecyclePolicyText: &policyText, RepositoryName: &repository}); err != nil {
				return err
			}
			return nil
		}
	}
	return nil
}