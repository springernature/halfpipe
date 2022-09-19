package dependabot

import (
	"github.com/sirupsen/logrus"
	"path/filepath"
	"strings"
)

type Filter interface {
	Filter(paths []string) MatchedPaths
}

type filter struct {
	skipEcosystems []string
	supportedFiles map[string]string
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

func (f filter) shouldInclude(path string) (bool, string) {
	fileName := filepath.Base(path)
	if ecosystem, ok := f.supportedFiles[fileName]; ok && !f.shouldFilterOutEcosystem(path, ecosystem) {
		return true, ecosystem
	}
	return false, ""
}

func (f filter) Filter(paths []string) MatchedPaths {
	filtered := MatchedPaths{}
	addedActions := false
	for _, path := range paths {
		if include, ecosystem := f.shouldInclude(path); include {
			filtered[path] = ecosystem
		}

		if strings.HasPrefix(path, ".github/workflows") && !addedActions {
			filtered["/"] = "github-actions"
			addedActions = true
		}
	}
	return filtered
}

func NewFilter(skipEcosystems []string) Filter {
	return filter{
		skipEcosystems: skipEcosystems,
		supportedFiles: map[string]string{
			"Dockerfile":        "docker",
			"package-lock.json": "npm",
			"yarn.lock":         "npm",
			"Gemfile.lock":      "bundler",
			"pom.xml":           "maven",
			"build.gradle":      "gradle",
			"build.gradle.kt":   "gradle",
			"go.mod":            "gomod",
		},
	}
}
