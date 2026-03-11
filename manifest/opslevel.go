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

// ParseOpsLevel reads and parses opslevel.yml from the given directory.
// Returns the parsed OpsLevel, a boolean indicating whether the file was found,
// and an error if the file exists but could not be parsed.
func ParseOpsLevel(fs afero.Afero, dir string) (OpsLevel, bool, error) {
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
