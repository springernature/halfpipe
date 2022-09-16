package dependabot

import (
	"github.com/sirupsen/logrus"
	"path/filepath"
	"strings"
)

type Filter interface {
	Filter(paths []string, skipEcosystems []string) []string
}

type filter struct {
}

func (f filter) shouldFilterOutEcosystem(path string, ecosystem string, skipEcosystems []string) bool {
	for _, skipEcosystem := range skipEcosystems {
		if skipEcosystem == ecosystem {
			logrus.Debugf("Removing '%s' due to filtered out ecosystem '%s'", path, ecosystem)
			return true
		}
	}
	return false
}

func (f filter) shouldInclude(path string, skipEcosystems []string) bool {
	fileName := filepath.Base(path)
	if ecosystem, ok := SupportedFiles[fileName]; ok && !f.shouldFilterOutEcosystem(path, ecosystem, skipEcosystems) {
		return true
	}
	return false
}

func (f filter) Filter(paths []string, skipEcosystems []string) (filtered []string) {
	addedActions := false
	for _, path := range paths {
		if f.shouldInclude(path, skipEcosystems) {
			filtered = append(filtered, path)
		}

		if strings.HasPrefix(path, ".github/workflows") && !addedActions {
			filtered = append(filtered, "github-actions")
			addedActions = true
		}
	}
	return
}

func NewFilter() Filter {
	return filter{}
}
