package dependabot

import (
	"github.com/sirupsen/logrus"
	"path/filepath"
	"strings"
)

type Filter interface {
	Filter(paths []string) []string
}

type filter struct {
	skipEcosystems []string
}

func (f filter) shouldFilterOutEcosystem(path string, ecosystem string) bool {
	for _, skipEcosystem := range f.skipEcosystems {
		if skipEcosystem == ecosystem {
			logrus.Debugf("Removing '%s' due to filtered out ecosystem '%s'", path, ecosystem)
			return true
		}
	}
	return false
}

func (f filter) shouldInclude(path string) bool {
	fileName := filepath.Base(path)
	if ecosystem, ok := SupportedFiles[fileName]; ok && !f.shouldFilterOutEcosystem(path, ecosystem) {
		return true
	}
	return false
}

func (f filter) Filter(paths []string) (filtered []string) {
	addedActions := false
	for _, path := range paths {
		if f.shouldInclude(path) {
			filtered = append(filtered, path)
		}

		if strings.HasPrefix(path, ".github/workflows") && !addedActions {
			filtered = append(filtered, "github-actions")
			addedActions = true
		}
	}
	return
}

func NewFilter(skipEcosystems []string) Filter {
	return filter{skipEcosystems}
}
