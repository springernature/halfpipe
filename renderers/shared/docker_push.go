package shared

import (
	"fmt"
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
	"strings"
)

func CachePath(task manifest.DockerPush, tag string) string {
	image, _ := SplitTag(task.Image)

	if strings.HasPrefix(task.Image, config.DockerRegistry) {
		r := strings.Replace(image, config.DockerRegistry, fmt.Sprintf("%scache/", config.DockerRegistry), 1)
		return r + tag
	} else {
		return config.DockerRegistry + "cache/" + image + tag
	}
}

func SplitTag(image string) (string, string) {
	split := strings.Split(image, ":")
	if len(split) == 2 {
		return split[0], split[1]
	}
	return split[0], ""
}
