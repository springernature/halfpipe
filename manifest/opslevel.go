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
// The returned OpsLevel has RelativePath set to the path of the file relative to
// gitRootDir (or startDir if gitRootDir is empty), and ParseError set if the file
// was found but could not be parsed.
func ParseOpsLevel(fs afero.Afero, startDir string, gitRootDir string) OpsLevel {
	dir := startDir
	for {
		opsLevel, found, err := parseOpsLevelInDir(fs, dir)
		if err != nil {
			return OpsLevel{RelativePath: opsLevelRelativePath(dir, startDir), ParseError: err.Error()}
		}
		if found {
			opsLevel.RelativePath = opsLevelRelativePath(dir, startDir)
			return opsLevel
		}

		if gitRootDir == "" || filepath.Clean(dir) == filepath.Clean(gitRootDir) {
			break
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return OpsLevel{}
}

func opsLevelRelativePath(dir, startDir string) string {
	rel, err := filepath.Rel(startDir, dir)
	if err != nil {
		rel = dir
	}
	return filepath.Join(rel, "opslevel.yml")
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
