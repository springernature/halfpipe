package shared

import (
	"fmt"
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
	"path"
	"strings"
)

func CachePath(task manifest.DockerPush, tag string) string {
	image, _ := SplitTag(task.Image)
	if tag != "" && !strings.HasPrefix(tag, ":") {
		tag = fmt.Sprintf(":%s", tag)
	}

	if strings.HasPrefix(task.Image, config.DockerRegistry) {
		r := strings.Replace(image, config.DockerRegistry, path.Join(config.DockerRegistry, "cache"), 1)
		return r + tag
	} else {
		return path.Join(config.DockerRegistry, "cache", image) + tag
	}
}

func SplitTag(image string) (string, string) {
	split := strings.Split(image, ":")
	if len(split) == 2 {
		return split[0], split[1]
	}
	return split[0], ""
}
