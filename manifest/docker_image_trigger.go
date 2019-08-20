package manifest

import "strings"

type DockerTrigger struct {
	Type  string
	Image string `json:"image,omitempty" yaml:"image,omitempty"`
}

func (d DockerTrigger) GetTriggerName() string {
	imageName := d.Image
	parts := strings.Split(imageName, "/")
	return parts[len(parts)-1]
}
