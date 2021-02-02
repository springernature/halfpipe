package actions

import (
	"fmt"
	"github.com/springernature/halfpipe/manifest"
	"strings"
)

func (a *Actions) runSteps(task manifest.Run) Steps {
	steps := Steps{}

	run := Step{
		Name: "run",
		Env:  Env(task.Vars),
	}
	if len(run.Env) == 0 {
		run.Env = make(map[string]string)
	}

	prefix := ""
	if a.workingDir != "" {
		prefix = fmt.Sprintf("cd %s;", a.workingDir)
	}
	script := task.Script
	if !strings.HasPrefix(script, "./") && !strings.HasPrefix(script, "/") && !strings.HasPrefix(script, `\`) {
		script = "./" + script
	}
	script = strings.Replace(script, `"`, `\"`, -1)

	if task.Docker.Image != "" {
		run.Uses = "docker://" + task.Docker.Image
		run.With = With{
			{"entrypoint", "/bin/sh"},
			{"args", fmt.Sprintf(`-c "%s %s"`, prefix, script)},
		}
	} else {
		run.Run = task.Script
	}

	steps = append(steps, dockerLogin(task.Docker.Image, task.Docker.Username, task.Docker.Password)...)
	steps = append(steps, run)
	return steps
}
