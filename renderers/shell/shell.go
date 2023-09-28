package shell

import (
	"fmt"
	"github.com/springernature/halfpipe"
	"github.com/springernature/halfpipe/manifest"
	"strings"
)

type shell struct {
	taskName string
}

func NewShell(taskName string) halfpipe.Renderer {
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

	return "", fmt.Errorf("task not found with name '%s' and type 'run' or 'docker-compose'", s.taskName)
}

func renderRunCommand(task manifest.Run, team string) string {
	s := []string{
		"docker run -it",
		`-v "$PWD":/app`,
		"-w /app",
	}

	for k, v := range task.Vars {
		s = append(s, fmt.Sprintf("-e %s=%s", k, vaultLookup(v, team)))
	}

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

	for k, v := range task.Vars {
		s = append(s, fmt.Sprintf("-e %s=%s", k, vaultLookup(v, team)))
	}

	s = append(s, "--use-aliases", task.Service)

	if task.Command != "" {
		s = append(s, task.Command)
	}

	return strings.Join(s, " \\ \n  ")
}

func vaultLookup(s string, team string) string {
	if !isSecret(s) {
		return s
	}
	s = strings.TrimSpace(s[2 : len(s)-2])

	if isKeyValueSecret(s) {
		parts := strings.Split(s, ".")
		vaultFolder := team
		if isShared(parts[0]) {
			vaultFolder = "shared"
		}
		return fmt.Sprintf("$(vault kv get -field=%s /springernature/%s/%s)", parts[1], vaultFolder, parts[0])
	}

	if isAbsolutePathSecret(s) {
		parts := strings.Split(s, " ")
		return fmt.Sprintf("$(vault kv get -field=%s /springernature/%s/%s)", parts[1], team, parts[0])
	}

	return s
}

// **************************************************
// all this copied from renderers/actions/secrets.go

// check if a secret matches one of the shared secrets
// vault kv list /springernature/shared
func isShared(s string) bool {
	return map[string]bool{
		"PPG-gradle-version-reporter":         true,
		"PPG-owasp-dependency-reporter":       true,
		"artifactory":                         true,
		"artifactory-support":                 true,
		"artifactory_test":                    true,
		"bla":                                 true,
		"burpsuiteenterprise":                 true,
		"content_hub-casper-credentials-live": true,
		"content_hub-casper-credentials-qa":   true,
		"contrastsecurity":                    true,
		"eas-sigrid":                          true,
		"ee-sso-route-service":                true,
		"fastly":                              true,
		"grafana":                             true,
		"halfpipe-artifacts":                  true,
		"halfpipe-docker-config":              true,
		"halfpipe-gcr":                        true,
		"halfpipe-github":                     true,
		"halfpipe-ml-deploy":                  true,
		"halfpipe-semver":                     true,
		"halfpipe-slack":                      true,
		"katee-tls-dev":                       true,
		"katee-tls-prod":                      true,
		"sentry-release-integration":          true,
	}[s]
}

func isSecret(s string) bool {
	return strings.HasPrefix(s, "((") && strings.HasSuffix(s, "))")
}

func isAbsolutePathSecret(s string) bool {
	return len(strings.Split(s, " ")) == 2
}

func isKeyValueSecret(s string) bool {
	return len(strings.Split(s, ".")) == 2
}
