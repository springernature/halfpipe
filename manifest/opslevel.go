package manifest

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
)

type opsLevelFile struct {
	Component OpsLevel `yaml:"component"`
}

// ParseOpsLevel searches for opslevel.yml starting from startDir and walking up
// to gitRootDir (inclusive). It returns the first opslevel.yml found (closest wins).
// If gitRootDir is empty, only startDir is checked.
// Returns the parsed OpsLevel, a boolean indicating whether the file was found,
// and an error if the file exists but could not be parsed.
func ParseOpsLevel(fs afero.Afero, startDir string, gitRootDir string) (OpsLevel, bool, error) {
	dir := startDir
	for {
		opsLevel, found, err := parseOpsLevelInDir(fs, dir)
		if found || err != nil {
			return opsLevel, found, err
		}

		if gitRootDir == "" || filepath.Clean(dir) == filepath.Clean(gitRootDir) {
			break
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root without finding gitRootDir
			break
		}
		dir = parent
	}

	return OpsLevel{}, false, nil
}

func parseOpsLevelInDir(fs afero.Afero, dir string) (OpsLevel, bool, error) {
	opsLevelPath := filepath.Join(dir, "opslevel.yml")

	data, err := fs.ReadFile(opsLevelPath)
	if err != nil {
		if os.IsNotExist(err) {
			return OpsLevel{}, false, nil
		}
		return OpsLevel{}, false, fmt.Errorf("failed to read opslevel.yml: %w", err)
	}

	var file opsLevelFile
	if err := yaml.Unmarshal(data, &file); err != nil {
		return OpsLevel{}, true, fmt.Errorf("failed to parse opslevel.yml: %w", err)
	}

	return file.Component, true, nil
}
