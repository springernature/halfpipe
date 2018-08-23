package project

import (
	"bytes"
	"io/ioutil"
	"os/exec"
	"strings"
)

/*
HERE BE DRAGONS! UNTESTED DRAGONS!

So, we cannot simply do `git rev-parse --abbrev-ref HEAD` as upload might be run in Concourse

In Concourse we are in detached HEAD state branch so the above would just yield `HEAD`

`git branch` will yield
* (HEAD detached at short-sha)
  master
`
if resource is set to another branch, master will be that branch

So, we need to do some check to determine if we are in concourse or on a developer machine
If `git branch` yields two line where the first line contains `HEAD detached at`
we assume that we are running in Concourse, thus getting the second line from `git branch` will give us the branch.

If the above check are untrue we can assume that we are on a developer machine and
`git rev-parse --abbrev-ref HEAD` should work.

*/
type GitBranchResolver func() (branch string, err error)

func runGitBranch() (output []string, err error) {
	if _, e := exec.LookPath("git"); e != nil {
		err = ErrGitNotFound
		return
	}

	var stdout bytes.Buffer
	cmd := exec.Command("git", "branch")
	cmd.Stdout = &stdout
	cmd.Stderr = ioutil.Discard

	if runErr := cmd.Run(); runErr != nil {
		err = runErr
		return
	}

	output = strings.Split(strings.TrimSpace(stdout.String()), "\n")

	return
}

func runGitRevParse() (output []string, err error) {
	var stdout bytes.Buffer
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD") // nolint
	cmd.Stdout = &stdout
	cmd.Stderr = ioutil.Discard

	if runErr := cmd.Run(); runErr != nil {
		err = runErr
		return
	}

	output = strings.Split(strings.TrimSpace(stdout.String()), "\n")
	return
}

func runningInConcourse(gitBranchOutput []string, gitRevParseOutput []string) bool {
	if len(gitBranchOutput) == 2 {
		if strings.Contains(gitBranchOutput[0], "HEAD detached at") {
			if len(gitRevParseOutput) == 1 && gitRevParseOutput[0] == "HEAD" {
				return true
			}
		}
	}
	return false
}

func BranchResolver() (branch string, err error) {
	gitBranchOutput, err := runGitBranch()
	if err != nil {
		panic(err)
	}

	gitRevParseOutput, err := runGitRevParse()
	if err != nil {
		panic(err)
	}

	if runningInConcourse(gitBranchOutput, gitRevParseOutput) {
		branch = strings.TrimSpace(gitBranchOutput[1])
	} else {
		branch = gitRevParseOutput[0]
	}
	return
}
