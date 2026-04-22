package manifest

import "strings"

// docker trigger runs the pipeline when a docker image has been updated.
type DockerTrigger struct {
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
	// Docker image to watch for updates.
	Image string `json:"image,omitempty" yaml:"image,omitempty" jsonschema:"required"`
	// Username for private Docker registries.
	Username string `json:"username,omitempty" yaml:"username,omitempty" secretAllowed:"true"`
	// Password for private Docker registries.
	Password string `json:"password,omitempty" yaml:"password,omitempty" secretAllowed:"true"`
}

func (d DockerTrigger) GetTriggerAttempts() int {
	return 2
}

func (d DockerTrigger) MarshalYAML() (any, error) {
	d.Type = "docker"
	return d, nil
}

func (d DockerTrigger) GetTriggerName() string {
	/*
		Name components may contain lowercase letters, digits and separators.
		A separator is defined as a period, one or two underscores, or one or more dashes.
		A name component may not start or end with a separator

		A tag name must be valid ASCII and may contain lowercase and uppercase letters, digits, underscores, periods and dashes.
		A tag name may not start with a period or a dash and may contain a maximum of 128 characters.

		https://docs.docker.com/engine/reference/commandline/tag/
	*/

	parts := strings.Split(d.Image, "/")
	name := parts[len(parts)-1]
	lowerCasedName := strings.ToLower(name)
	withoutUnderscore := strings.Replace(lowerCasedName, "_", "-", -1)
	replaceColonWithDot := strings.Replace(withoutUnderscore, ":", ".", -1)
	return replaceColonWithDot
}
