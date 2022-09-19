package dependabot

import (
	"fmt"
	"path/filepath"
	"sort"
)

func renderPath(path string, ecosystem string) Dependency {
	dir := filepath.Dir(path)
	if dir == "." {
		dir = "/"
	}
	if dir != "/" {
		dir = fmt.Sprintf("/%s", dir)
	}

	return Dependency{
		PackageEcosystem: ecosystem,
		Directory:        dir,
		Schedule:         Schedule{Interval: "daily"},
	}
}

func Render(matchedPaths MatchedPaths) Config {
	paths := []string{}
	for path := range matchedPaths {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	updates := []Dependency{}
	for _, path := range paths {
		updates = append(updates, renderPath(path, matchedPaths[path]))
	}
	return Config{
		Version: 2,
		Updates: updates,
	}
}
