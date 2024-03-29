package stack

import (
	"github.com/stretchr/testify/assert"
	"github.com/sparetimecoders/build-tools/pkg/templating"
	"testing"
)

func TestNone_Scaffold(t *testing.T) {
	stack := &None{}

	err := stack.Scaffold("dir", templating.TemplateData{
		ProjectName:    "test",
		Badges:         nil,
		Organisation:   "org.example",
		RepositoryUrl:  "git@github.com/org/example",
		RepositoryHost: "github.com",
		RepositoryPath: "/org/example",
	})

	assert.NoError(t, err)
}

func TestNone_Name(t *testing.T) {
	stack := &None{}

	assert.Equal(t, "none", stack.Name())
}
