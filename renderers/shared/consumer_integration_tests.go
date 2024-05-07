package shared

import (
	"fmt"
	"sort"
	"strings"
)

type CacheDirs struct {
	RunnerDir    string
	ContainerDir string
}

func ConsumerIntegrationTestScript(keys []string, cacheDirs []CacheDirs, isConcourse bool) string {
	var envStrings []string
	for _, key := range keys {
		envStrings = append(envStrings, fmt.Sprintf("-e %s", key))
	}
	sort.Strings(envStrings)
	envOption := strings.Join(envStrings, " ")

	var volumeFlags []string
	for _, cache := range cacheDirs {
		volumeFlags = append(volumeFlags, fmt.Sprintf("-v %s:%s", cache.RunnerDir, cache.ContainerDir))
	}

	if isConcourse {
		// For now we only pass on the docker.sock to CDCs when running in Concourse
		volumeFlags = append(volumeFlags, fmt.Sprintf("-v %s:%s", "/var/run/docker.sock", "/var/run/docker.sock"))
	}
	volumeOption := strings.Join(volumeFlags, " ")

	return fmt.Sprintf(`export ENV_OPTIONS="%s"
export VOLUME_OPTIONS="%s"
run-cdc.sh`, envOption, volumeOption)
}
