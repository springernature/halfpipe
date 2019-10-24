package mapper

import (
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/manifest"
)

type Mapper interface {
	Apply(original manifest.Manifest) (updated manifest.Manifest)
}

type mapper struct {
	mappers []Mapper
}

func (m mapper) Apply(original manifest.Manifest) (updated manifest.Manifest) {
	updated = original

	for _, mm := range m.mappers {
		updated = mm.Apply(updated)
	}

	return updated
}

func New() Mapper {
	return mapper{
		mappers: []Mapper{
			NewNotificationsMapper(),
			NewDockerComposeMapper(afero.Afero{Fs: afero.NewOsFs()}),
		},
	}
}
