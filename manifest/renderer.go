package manifest

import "github.com/simonjohansson/yaml"

func Render(manifest Manifest) (y []byte, err error) {
	return yaml.Marshal(manifest)
}
