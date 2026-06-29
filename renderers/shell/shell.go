package shell

import (
	"fmt"
	"sort"
	"strings"

	"github.com/springernature/halfpipe"
	"github.com/springernature/halfpipe/manifest"
	"github.com/springernature/halfpipe/renderers/shared/secrets"
)

type shell struct {
	taskName string
}

func New(taskName string) halfpipe.Renderer {
	return shell{taskName: taskName}
}

func (s shell) Render(man manifest.Manifest) (string, error) {
	task := man.Tasks.GetTask(s.taskName)

	switch t := task.(type) {
	case manifest.Run:
		return renderRunCommand(t, man.Team), nil
	case manifest.DockerCompose:
		return renderDockerComposeCommand(t, man.Team), nil
	case manifest.Buildpack:
		return renderBuildpackCommand(t, man.Team), nil
	}

	availableTasks := []string{}
	for _, t := range man.Tasks.Flatten() {
		switch t := t.(type) {
		case manifest.Run, manifest.DockerCompose, manifest.Buildpack:
			availableTasks = append(availableTasks, fmt.Sprintf("  %s", t.GetName()))
		}
	}

	return "", fmt.Errorf("task not found named '%s'. Supported task types for exec command are run, docker-compose and buildpack.\n\navailable tasks:\n%s",
		s.taskName,
		strings.Join(availableTasks, "\n"),
	)
}

func renderRunCommand(task manifest.Run, team string) string {
	s := []string{
		"docker run -it --rm",
		`-v "$PWD":/app`,
		"-w /app",
	}

	vars := []string{}
	for k, v := range task.Vars {
		vars = append(vars, fmt.Sprintf("-e %s=%s", k, quoteValue(v, team)))
	}
	sort.Strings(vars)
	s = append(s, vars...)

	s = append(s, task.Docker.Image, task.Script)

	return strings.Join(s, " \\ \n  ")
}

func quoteValue(v string, team string) string {
	secret := secrets.New(v, team)
	if secret != nil {
		field := secret.Key
		path := strings.Replace(secret.MapPath, "/springernature/data/", "/springernature/", 1)
		return fmt.Sprintf("\"$(vault kv get -field=%s %s)\"", field, path)
	}
	return fmt.Sprintf("'%s'", v)
}

func renderDockerComposeCommand(task manifest.DockerCompose, team string) string {
	s := []string{"docker compose"}
	s = append(s, toMultipleArgs("-f", task.ComposeFiles)...)
	s = append(s,
		"run",
		"--rm",
		`-v "$PWD":/app`,
		"-w /app",
	)

	vars := []string{}
	for k, v := range task.Vars {
		vars = append(vars, fmt.Sprintf("-e %s=%s", k, quoteValue(v, team)))
	}
	sort.Strings(vars)
	s = append(s, vars...)

	s = append(s, "--use-aliases", task.Service)

	if task.Command != "" {
		s = append(s, task.Command)
	}

	return strings.Join(s, " \\ \n  ")
}

func toMultipleArgs(flag string, args []string) []string {
	out := []string{}
	for _, arg := range args {
		out = append(out, fmt.Sprintf("%s %s", flag, arg))
	}
	return out
}

func renderBuildpackCommand(task manifest.Buildpack, team string) string {
	path := "."
	if task.Path != "" {
		path = task.Path
	}

	s := []string{
		fmt.Sprintf("pack build %s", task.Image),
		fmt.Sprintf("--path %s", path),
		fmt.Sprintf("--builder %s", task.Builder),
	}

	for _, bp := range task.Buildpacks {
		s = append(s, fmt.Sprintf("--buildpack %s", bp))
	}

	vars := []string{}
	for k, v := range task.Vars {
		vars = append(vars, fmt.Sprintf("--env %s=%s", k, quoteValue(v, team)))
	}
	sort.Strings(vars)
	s = append(s, vars...)

	s = append(s, fmt.Sprintf("--tag %s:local", task.Image))
	s = append(s, "--trust-builder")
	s = append(s, "--platform linux/amd64")

	return strings.Join(s, " \\\n  ")
}
