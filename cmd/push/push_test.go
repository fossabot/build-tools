package main

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestPush_BadDockerHost(t *testing.T) {
	oldPwd, _ := os.Getwd()
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer os.RemoveAll(name)
	err := os.Chdir(name)
	assert.NoError(t, err)
	defer os.Chdir(oldPwd)

	os.Clearenv()
	_ = os.Setenv("DOCKER_HOST", "abc-123")
	code := doPush()
	assert.Equal(t, -1, code)
}

func TestPush(t *testing.T) {
	oldPwd, _ := os.Getwd()
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer os.RemoveAll(name)
	err := os.Chdir(name)
	assert.NoError(t, err)
	defer os.Chdir(oldPwd)
	os.Clearenv()
	code := doPush()
	assert.Equal(t, -2, code)
}
