package config

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	oldPwd, tempDir := setup()
	code := m.Run()
	teardown(oldPwd, tempDir)
	os.Exit(code)
}

func setup() (string, string) {
	oldPwd, _ := os.Getwd()
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	_ = os.Chdir(name)

	return oldPwd, name
}

func teardown(oldPwd, tempDir string) {
	_ = os.RemoveAll(tempDir)
	_ = os.Chdir(oldPwd)
}

func TestIdentify_Azure(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("VSTS_PROCESS_LOOKUP_ID", "1")
	_ = os.Setenv("BUILD_SOURCEVERSION", "abc123")
	_ = os.Setenv("BUILD_REPOSITORY_NAME", "reponame")
	_ = os.Setenv("BUILD_SOURCEBRANCHNAME", "feature/first test")

	cfg, err := Load(".")
	assert.NoError(t, err)
	result := cfg.CurrentCI()
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "abc123", result.Commit())
	assert.Equal(t, "reponame", result.BuildName())
	assert.Equal(t, "feature/first test", result.Branch())
	assert.Equal(t, "feature_first_test", result.BranchReplaceSlash())
}

func TestBranch_VCS_Fallback_Azure(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("CI", "azure")

	dir, _ := ioutil.TempDir("", "build-tools")
	defer os.RemoveAll(dir)

	InitRepoWithCommit(dir)

	cfg, err := Load(dir)
	assert.NoError(t, err)
	result := cfg.CurrentCI()
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "master", result.Branch())
}

func TestCommit_VCS_Fallback_Azure(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("CI", "azure")

	dir, _ := ioutil.TempDir("", "build-tools")
	defer os.RemoveAll(dir)

	hash, _ := InitRepoWithCommit(dir)

	cfg, err := Load(dir)
	assert.NoError(t, err)
	result := cfg.CurrentCI()
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, hash.String(), result.Commit())
}
