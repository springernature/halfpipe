package manifest

import (
	"gopkg.in/yaml.v3"
)

type Metadata struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
}
type VelaManifest struct {
	Kind     string `yaml:"kind"`
	Metadata Metadata
}

func UnmarshalVelaManifest(bytes []byte) (vm VelaManifest, err error) {
	err = yaml.Unmarshal(bytes, &vm)
	if err != nil {
		return vm, err
	}
	return vm, nil
}
