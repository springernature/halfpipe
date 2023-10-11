package shell

import (
	"fmt"
	"github.com/springernature/halfpipe"
	"github.com/springernature/halfpipe/manifest"
	"github.com/springernature/halfpipe/renderers/shared/secrets"
	"sort"
	"strings"
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
	}

	errMsg := "task not found with name '%s' and type 'run' or 'docker-compose\n\navailable tasks:\n"
	for _, t := range man.Tasks.Flatten() {
		switch t := t.(type) {
		case manifest.Run, manifest.DockerCompose:
			errMsg += fmt.Sprintf("  %s", t.GetName())
		}
	}
	return "", fmt.Errorf(errMsg, s.taskName)
}

func renderRunCommand(task manifest.Run, team string) string {
	s := []string{
		"docker run -it",
		`-v "$PWD":/app`,
		"-w /app",
	}

	vars := []string{}
	for k, v := range task.Vars {
		vars = append(vars, fmt.Sprintf(`-e %s="%s"`, k, convertSecret(v, team)))
	}
	sort.Strings(vars)
	s = append(s, vars...)

	s = append(s, task.Docker.Image, task.Script)

	return strings.Join(s, " \\ \n  ")
}

func renderDockerComposeCommand(task manifest.DockerCompose, team string) string {
	s := []string{
		"docker compose",
		fmt.Sprintf("-f %s", task.ComposeFile),
		"run",
		`-v "$PWD":/app`,
		"-w /app",
	}

	vars := []string{}
	for k, v := range task.Vars {
		vars = append(vars, fmt.Sprintf(`-e %s="%s"`, k, convertSecret(v, team)))
	}
	sort.Strings(vars)
	s = append(s, vars...)

	s = append(s, "--use-aliases", task.Service)

	if task.Command != "" {
		s = append(s, task.Command)
	}

	return strings.Join(s, " \\ \n  ")
}

func convertSecret(s string, team string) string {
	secret := secrets.New(s, team)
	if secret == nil {
		return s
	}
	return fmt.Sprintf("$(vault kv get -field=%s /springernature/%s)", secret.Key, secret.MapPath)
}
