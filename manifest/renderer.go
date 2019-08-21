package manifest

import "gopkg.in/yaml.v2"

func Render(manifest Manifest) (y []byte, err error) {
	return yaml.Marshal(manifest)
}
