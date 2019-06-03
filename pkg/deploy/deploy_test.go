package deploy

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"gitlab.com/sparetimecoders/build-tools/pkg/kubectl"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestDeploy_MissingDeploymentFilesDir(t *testing.T) {
	client := &kubectl.MockKubectl{
		Responses: []error{nil},
	}
	defer client.Cleanup()

	err := Deploy(".", "abc123", "image", "20190513-17:22:36", client)

	assert.EqualError(t, err, "open deployment_files: no such file or directory")
}

func TestDeploy_NoFiles(t *testing.T) {
	client := &kubectl.MockKubectl{
		Responses: []error{nil},
	}

	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer os.RemoveAll(name)
	_ = os.Mkdir(filepath.Join(name, "deployment_files"), 0777)

	err := Deploy(name, "abc123", "image", "20190513-17:22:36", client)

	assert.NoError(t, err)
	assert.Equal(t, 0, len(client.Inputs))
}

func TestDeploy_NoEnvSpecificFiles(t *testing.T) {
	client := &kubectl.MockKubectl{
		Responses: []error{nil},
	}

	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer os.RemoveAll(name)
	_ = os.Mkdir(filepath.Join(name, "deployment_files"), 0777)
	yaml := `
apiVersion: v1
kind: Namespace
metadata:
  name: dummy
`
	deployFile := filepath.Join(name, "deployment_files", "deploy.yaml")
	_ = ioutil.WriteFile(deployFile, []byte(yaml), 0777)

	err := Deploy(name, "abc123", "image", "20190513-17:22:36", client)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(client.Inputs))
	assert.Equal(t, yaml, client.Inputs[0])
}

func TestDeploy_UnreadableFile(t *testing.T) {
	client := &kubectl.MockKubectl{
		Responses: []error{nil},
	}

	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer os.RemoveAll(name)
	_ = os.MkdirAll(filepath.Join(name, "deployment_files", "deploy.yaml"), 0777)

	err := Deploy(name, "abc123", "image", "20190513-17:22:36", client)

	assert.EqualError(t, err, fmt.Sprintf("read %s/deployment_files/deploy.yaml: is a directory", name))
}

func TestDeploy_FileBrokenSymlink(t *testing.T) {
	client := &kubectl.MockKubectl{
		Responses: []error{nil},
	}

	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer os.RemoveAll(name)
	_ = os.MkdirAll(filepath.Join(name, "deployment_files"), 0777)
	deployFile := filepath.Join(name, "deployment_files", "ns.yaml")
	_ = ioutil.WriteFile(deployFile, []byte("test"), 0777)
	_ = os.Symlink(deployFile, filepath.Join(name, "deployment_files", "deploy.yaml"))
	_ = os.Remove(deployFile)

	err := Deploy(name, "abc123", "image", "20190513-17:22:36", client)

	assert.EqualError(t, err, fmt.Sprintf("open %s/deployment_files/deploy.yaml: no such file or directory", name))
}

func TestDeploy_EnvSpecificFilesInSubDirectory(t *testing.T) {
	client := &kubectl.MockKubectl{
		Responses: []error{nil},
	}

	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer os.RemoveAll(name)
	_ = os.MkdirAll(filepath.Join(name, "deployment_files", "dummy"), 0777)
	yaml := `
apiVersion: v1
kind: Namespace
metadata:
  name: dummy
`
	deployFile := filepath.Join(name, "deployment_files", "dummy", "ns.yaml")
	_ = ioutil.WriteFile(deployFile, []byte(yaml), 0777)

	err := Deploy(name, "abc123", "image", "20190513-17:22:36", client)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(client.Inputs))
	assert.Equal(t, yaml, client.Inputs[0])
}

func TestDeploy_EnvSpecificFilesWithSuffix(t *testing.T) {
	client := &kubectl.MockKubectl{
		Responses: []error{nil},
	}

	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer os.RemoveAll(name)
	_ = os.Mkdir(filepath.Join(name, "deployment_files"), 0777)
	yaml := `
apiVersion: v1
kind: Namespace
metadata:
  name: dummy
`
	deployFile := filepath.Join(name, "deployment_files", "ns-dummy.yaml")
	_ = ioutil.WriteFile(deployFile, []byte(yaml), 0777)

	err := Deploy(name, "abc123", "image", "20190513-17:22:36", client)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(client.Inputs))
	assert.Equal(t, yaml, client.Inputs[0])
}

func TestDeploy_EnvSpecificFiles(t *testing.T) {
	client := &kubectl.MockKubectl{
		Responses: []error{nil},
	}

	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer os.RemoveAll(name)
	_ = os.MkdirAll(filepath.Join(name, "deployment_files", "prod"), 0777)
	yaml := `
apiVersion: v1
kind: Namespace
metadata:
  name: dummy
`
	deployFile := filepath.Join(name, "deployment_files", "ns-dummy.yaml")
	_ = ioutil.WriteFile(deployFile, []byte(yaml), 0777)
	_ = ioutil.WriteFile(filepath.Join(name, "deployment_files", "ns-prod.yaml"), []byte(yaml), 0777)
	_ = ioutil.WriteFile(filepath.Join(name, "deployment_files", "other-dummy.sh"), []byte(yaml), 0777)

	err := Deploy(name, "abc123", "image", "20190513-17:22:36", client)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(client.Inputs))
	assert.Equal(t, yaml, client.Inputs[0])
}

func TestDeploy_ErrorFromApply(t *testing.T) {
	client := &kubectl.MockKubectl{
		Responses: []error{errors.New("apply failed")},
	}

	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer os.RemoveAll(name)
	_ = os.Mkdir(filepath.Join(name, "deployment_files"), 0777)
	yaml := `
apiVersion: v1
kind: Namespace
metadata:
  name: dummy
`
	deployFile := filepath.Join(name, "deployment_files", "deploy.yaml")
	_ = ioutil.WriteFile(deployFile, []byte(yaml), 0777)

	err := Deploy(name, "abc123", "image", "20190513-17:22:36", client)

	assert.EqualError(t, err, "apply failed")
}

func TestDeploy_ErrorFromApplyInSubDirectory(t *testing.T) {
	client := &kubectl.MockKubectl{
		Responses: []error{errors.New("apply failed")},
	}

	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer os.RemoveAll(name)
	_ = os.MkdirAll(filepath.Join(name, "deployment_files", "dummy"), 0777)
	yaml := `
apiVersion: v1
kind: Namespace
metadata:
  name: dummy
`
	deployFile := filepath.Join(name, "deployment_files", "dummy", "ns.yaml")
	_ = ioutil.WriteFile(deployFile, []byte(yaml), 0777)

	err := Deploy(name, "abc123", "image", "20190513-17:22:36", client)

	assert.EqualError(t, err, "apply failed")
}

func TestDeploy_ReplacingCommitAndTimestamp(t *testing.T) {
	client := &kubectl.MockKubectl{
		Responses: []error{nil},
	}

	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer os.RemoveAll(name)
	_ = os.Mkdir(filepath.Join(name, "deployment_files"), 0777)
	yaml := `
apiVersion: v1
kind: Namespace
metadata:
  name: dummy
  commit: ${COMMIT}
  timestamp: ${TIMESTAMP}
`
	deployFile := filepath.Join(name, "deployment_files", "deploy.yaml")
	_ = ioutil.WriteFile(deployFile, []byte(yaml), 0777)

	err := Deploy(name, "abc123", "image", "2019-05-13T17:22:36Z01:00", client)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(client.Inputs))
	expectedInput := `
apiVersion: v1
kind: Namespace
metadata:
  name: dummy
  commit: abc123
  timestamp: 2019-05-13T17:22:36Z01:00
`
	assert.Equal(t, expectedInput, client.Inputs[0])
}

func TestDeploy_DeploymentExists(t *testing.T) {
	client := &kubectl.MockKubectl{
		Responses:  []error{nil},
		Deployment: true,
	}

	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer os.RemoveAll(name)
	_ = os.Mkdir(filepath.Join(name, "deployment_files"), 0777)
	yaml := `
apiVersion: v1
kind: Namespace
metadata:
  name: dummy
`
	deployFile := filepath.Join(name, "deployment_files", "deploy.yaml")
	_ = ioutil.WriteFile(deployFile, []byte(yaml), 0777)

	err := Deploy(name, "abc123", "image", "20190513-17:22:36", client)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(client.Inputs))
	assert.Equal(t, yaml, client.Inputs[0])
}

func TestDeploy_RolloutStatusFail(t *testing.T) {
	client := &kubectl.MockKubectl{
		Responses:  []error{nil},
		Deployment: true,
		Status:     false,
	}

	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer os.RemoveAll(name)
	_ = os.Mkdir(filepath.Join(name, "deployment_files"), 0777)
	yaml := `
apiVersion: v1
kind: Namespace
metadata:
  name: dummy
`
	deployFile := filepath.Join(name, "deployment_files", "deploy.yaml")
	_ = ioutil.WriteFile(deployFile, []byte(yaml), 0777)

	err := Deploy(name, "abc123", "image", "20190513-17:22:36", client)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(client.Inputs))
	assert.Equal(t, yaml, client.Inputs[0])
}