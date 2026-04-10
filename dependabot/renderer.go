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

	usedRegistries := map[string]bool{}
	updates := []Dependency{}
	for _, name := range ecosystemNames {
		cfg := ecosystems[name]
		dirs := cfg.directories
		if dirs == nil {
			dirs = defaultDirectories
		}
		updates = append(updates, Dependency{
			PackageEcosystem:      name,
			Directories:           dirs,
			Schedule:              Schedule{Interval: "weekly"},
			Cooldown:              Cooldown{DefaultDays: 5},
			OpenPullRequestsLimit: 10,
			Labels:                []string{"dependencies", name},
			VersioningStrategy:    cfg.versioningStrategy,
			Groups:                cfg.groups,
			Registries:            cfg.registries,
		})
		for _, r := range cfg.registries {
			usedRegistries[r] = true
		}
	}

	var registries map[string]Registry
	if len(usedRegistries) > 0 {
		registries = map[string]Registry{}
		for name := range usedRegistries {
			registries[name] = registryDefinitions[name]
		}
	}

	return Config{
		Version:    2,
		Registries: registries,
		Updates:    updates,
	}
}
