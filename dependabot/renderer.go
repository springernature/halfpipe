package dependabot

import (
	"path/filepath"
	"sort"
)

func Render(matchedPaths MatchedPaths) Config {
	// Group discovered paths by ecosystem and derive directories.
	ecosystemDirs := map[string]map[string]bool{}
	for filePath, ecosystem := range matchedPaths {
		if ecosystemDirs[ecosystem] == nil {
			ecosystemDirs[ecosystem] = map[string]bool{}
		}
		dir := "/" + filepath.Dir(filePath)
		if dir == "/." || dir == "//" {
			dir = "/"
		}
		ecosystemDirs[ecosystem][dir] = true
	}

	ecosystemNames := make([]string, 0, len(ecosystemDirs))
	for name := range ecosystemDirs {
		ecosystemNames = append(ecosystemNames, name)
	}
	sort.Strings(ecosystemNames)

	updates := []Dependency{}
	for _, name := range ecosystemNames {
		cfg := ecosystems[name]

		dirSet := ecosystemDirs[name]
		dirs := make([]string, 0, len(dirSet))
		for d := range dirSet {
			dirs = append(dirs, d)
		}
		sort.Strings(dirs)
		updates = append(updates, Dependency{
			PackageEcosystem:      name,
			Directories:           dirs,
			Schedule:              Schedule{Interval: "weekly"},
			Cooldown:              Cooldown{DefaultDays: 5},
			OpenPullRequestsLimit: 10,
			Labels:                []string{"dependencies", name},
			CommitMessage:         CommitMessage{Prefix: "chore", Include: "scope"},
			VersioningStrategy:    cfg.versioningStrategy,
			Groups:                cfg.groups,
			Ignore:                cfg.ignore,
		})
	}

	return Config{
		Version: 2,
		Updates: updates,
	}
}
