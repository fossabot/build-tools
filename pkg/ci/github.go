package ci

import (
	"strings"
)

type Github struct {
	*Common
	CICommit     string `env:"GITHUB_SHA"`
	CIBuildName  string `env:"RUNNER_WORKSPACE"`
	CIBranchName string `env:"GITHUB_REF"`
}

var _ CI = &Github{}

func (c *Github) Name() string {
	return "Github"
}

func (c *Github) BranchReplaceSlash() string {
	return branchReplaceSlash(c.Branch())
}

func (c *Github) BuildName() string {
	return c.Common.BuildName(strings.TrimPrefix(c.CIBuildName, "/home/runner/work/"))
}

func (c *Github) Branch() string {
	return c.Common.Branch(strings.TrimPrefix(c.CIBranchName, "refs/heads/"))
}

func (c *Github) Commit() string {
	return c.Common.Commit(c.CICommit)
}

func (c *Github) Configured() bool {
	return c.CIBuildName != ""
}
