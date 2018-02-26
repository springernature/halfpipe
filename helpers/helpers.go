package helpers

import (
	"strings"

	"github.com/blang/semver"
	"github.com/springernature/halfpipe/cmd/config"
	"github.com/springernature/halfpipe/sync"
)

func SecretToMapAndKey(secret string) (string, string) {
	s := strings.Replace(strings.Replace(secret, "((", "", -1), "))", "", -1)
	parts := strings.Split(s, ".")
	mapName, keyName := parts[0], parts[1]
	return mapName, keyName
}

func GetVersion() (semver.Version, error) {
	if config.Version == "" {
		return sync.DevVersion, nil
	}
	version, err := semver.Make(config.Version)
	if err != nil {
		return semver.Version{}, err
	}
	return version, nil
}
