package main

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/sparetimecoders/build-tools/pkg"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestVersion(t *testing.T) {
	out = &bytes.Buffer{}
	version = "1.0.0"
	exitFunc = func(code int) {
		assert.Equal(t, 0, code)
	}
	os.Args = []string{"build","-version"}
	main()

	assert.Equal(t, "Version: 1.0.0, commit none, built at unknown\n", out.(*bytes.Buffer).String())
}

func TestDeploy_BrokenConfig(t *testing.T) {
	exitFunc = func(code int) {
		assert.Equal(t, -1, code)
	}

	oldPwd, _ := os.Getwd()
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	yaml := `ci: []
`
	_ = ioutil.WriteFile(filepath.Join(name, ".buildtools.yaml"), []byte(yaml), 0777)

	err := os.Chdir(name)
	assert.NoError(t, err)
	defer func() { _ = os.Chdir(oldPwd) }()

	os.Args = []string{"deploy", "dummy"}
	main()
}

func TestDeploy_MissingEnvironment(t *testing.T) {
	exitFunc = func(code int) {
		assert.Equal(t, -0, code)
	}

	oldPwd, _ := os.Getwd()
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()

	err := os.Chdir(name)
	assert.NoError(t, err)
	defer func() { _ = os.Chdir(oldPwd) }()

	os.Args = []string{"deploy", "dummy"}
	main()
}

func TestDeploy_NoCI(t *testing.T) {
	exitFunc = func(code int) {
		assert.Equal(t, -2, code)
	}

	oldPwd, _ := os.Getwd()
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	yaml := `
environments:
  dummy:
    context: missing
    namespace: none
`
	_ = ioutil.WriteFile(filepath.Join(name, ".buildtools.yaml"), []byte(yaml), 0777)

	err := os.Chdir(name)
	assert.NoError(t, err)
	defer func() { _ = os.Chdir(oldPwd) }()

	os.Args = []string{"deploy", "dummy"}
	main()
}

func TestDeploy_NoEnv(t *testing.T) {
	exitFunc = func(code int) {
		assert.Equal(t, -0, code)
	}

	defer pkg.SetEnv("CI_COMMIT_SHA", "abc123")()
	defer pkg.SetEnv("CI_PROJECT_NAME", "dummy")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "master")()
	oldPwd, _ := os.Getwd()
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	yaml := `
environments:
  - name: dummy
    context: missing
    namespace: none
`
	_ = ioutil.WriteFile(filepath.Join(name, ".buildtools.yaml"), []byte(yaml), 0777)

	err := os.Chdir(name)
	assert.NoError(t, err)
	defer func() { _ = os.Chdir(oldPwd) }()

	os.Args = []string{"deploy"}
	main()
}

func TestDeploy_NoOptions(t *testing.T) {
	exitFunc = func(code int) {
		assert.Equal(t, -3, code)
	}

	defer pkg.SetEnv("CI_COMMIT_SHA", "abc123")()
	defer pkg.SetEnv("CI_PROJECT_NAME", "dummy")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "master")()
	oldPwd, _ := os.Getwd()
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	yaml := `
environments:
  dummy:
    context: missing
    namespace: none
`
	_ = ioutil.WriteFile(filepath.Join(name, ".buildtools.yaml"), []byte(yaml), 0777)

	err := os.Chdir(name)
	assert.NoError(t, err)
	defer func() { _ = os.Chdir(oldPwd) }()

	os.Args = []string{"deploy", "dummy"}
	main()
}

func TestDeploy_ContextAndNamespaceSpecified(t *testing.T) {
	exitFunc = func(code int) {
		assert.Equal(t, -3, code)
	}

	defer pkg.SetEnv("CI_COMMIT_SHA", "abc123")()
	defer pkg.SetEnv("CI_PROJECT_NAME", "dummy")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "master")()
	oldPwd, _ := os.Getwd()
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	yaml := `
environments:
  dummy:
    context: missing
    namespace: none
`
	_ = ioutil.WriteFile(filepath.Join(name, ".buildtools.yaml"), []byte(yaml), 0777)

	err := os.Chdir(name)
	assert.NoError(t, err)
	defer func() { _ = os.Chdir(oldPwd) }()

	os.Args = []string{"deploy", "-c", "other", "-n", "dev", "dummy"}
	main()
}
