package shared

import (
	"fmt"
	"github.com/springernature/halfpipe/manifest"
	"strings"
)

func BuildTestRoute(task manifest.DeployCF) string {
	return fmt.Sprintf("%s-%s-CANDIDATE.%s", strings.Replace(task.CfApplication.Name, "_", "-", -1), strings.Replace(task.Space, "_", "-", -1), task.TestDomain)
}
