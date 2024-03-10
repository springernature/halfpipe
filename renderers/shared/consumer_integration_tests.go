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

func ConsumerIntegrationTestScript(keys []string, cacheDirs []CacheDirs) string {
	var envStrings []string
	for _, key := range keys {
		envStrings = append(envStrings, fmt.Sprintf("-e %s", key))
	}
	sort.Strings(envStrings)
	envOption := strings.Join(envStrings, " ")

	var cacheVolumeFlags []string
	for _, cache := range cacheDirs {
		cacheVolumeFlags = append(cacheVolumeFlags, fmt.Sprintf("-v %s:%s", cache.RunnerDir, cache.ContainerDir))
	}

	volumeOption := strings.Join(cacheVolumeFlags, " ")

	return fmt.Sprintf(`export ENV_OPTIONS="%s"
export VOLUME_OPTIONS="%s"
run-cdc.sh`, envOption, volumeOption)
}
