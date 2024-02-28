package shared

import (
	"fmt"
	"sort"
	"strings"
)

func ConsumerIntegrationTestScript(keys []string, cacheDirs []string) string {
	var envStrings []string
	for _, key := range keys {
		envStrings = append(envStrings, fmt.Sprintf("-e %s", key))
	}
	sort.Strings(envStrings)
	envOption := strings.Join(envStrings, " ")

	var cacheVolumeFlags []string
	for _, cacheVolume := range cacheDirs {
		cacheVolumeFlags = append(cacheVolumeFlags, fmt.Sprintf("-v %s:%s", cacheVolume, cacheVolume))
	}

	volumeOption := strings.Join(cacheVolumeFlags, " ")

	return fmt.Sprintf(`export ENV_OPTIONS="%s"
export VOLUME_OPTIONS="%s"
run-cdc.sh`, envOption, volumeOption)
}
