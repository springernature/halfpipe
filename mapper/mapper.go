package mapper

import (
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/manifest"
)

type Mapper interface {
	Apply(original manifest.Manifest) (updated manifest.Manifest, err error)
}

type mapper struct {
	mappers []Mapper
}

func (m mapper) Apply(original manifest.Manifest) (updated manifest.Manifest, err error) {
	updated = original

	for _, mm := range m.mappers {
		updated, err = mm.Apply(updated)
		if err != nil {
			return updated, err
		}
	}

	return updated, nil
}

func New() Mapper {
	return mapper{
		mappers: []Mapper{
			NewNotificationsMapper(),
			NewDockerComposeMapper(afero.Afero{Fs: afero.NewOsFs()}),
			NewCFDockerPushMapper(),
		},
	}
}
