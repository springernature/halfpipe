package dependabot

import (
	"sort"
)

var defaultDirectories = []string{"/**"}

func Render(matchedPaths MatchedPaths) Config {
	seen := map[string]bool{}
	ecosystemNames := []string{}
	for _, ecosystem := range matchedPaths {
		if !seen[ecosystem] {
			seen[ecosystem] = true
			ecosystemNames = append(ecosystemNames, ecosystem)
		}
	}
	sort.Strings(ecosystemNames)

	updates := []Dependency{}
	for _, name := range ecosystemNames {
		cfg := ecosystems[name]
		dirs := cfg.directories
		if dirs == nil {
			dirs = defaultDirectories
		}
		updates = append(updates, Dependency{
			PackageEcosystem:   name,
			Directories:        dirs,
			Schedule:           Schedule{Interval: "weekly"},
			Cooldown:           Cooldown{DefaultDays: 5},
			VersioningStrategy: cfg.versioningStrategy,
			Groups:             cfg.groups,
		})
	}
	return Config{
		Version: 2,
		Updates: updates,
	}
}
