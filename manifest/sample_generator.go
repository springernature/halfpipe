package manifest

import (
	"errors"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
)

var ErrHalfpipeAlreadyExists = errors.New("'.halfpipe.io' already exists")

type SampleGenerator interface {
	Generate() (err error)
}

type sampleGenerator struct {
	fs afero.Afero
}

func NewSampleGenerator(fs afero.Afero) SampleGenerator {
	return sampleGenerator{
		fs: fs,
	}
}

func (s sampleGenerator) Generate() (err error) {
	exists, err := s.fs.Exists(".halfpipe.io")
	if err != nil {
		return
	}

	if exists {
		err = ErrHalfpipeAlreadyExists
		return
	}

	manifest := Manifest{
		Team:     "CHANGE-ME",
		Pipeline: "CHANGE-ME",

		Tasks: []Task{
			Run{
				Type:   "run",
				Name:   "CHANGE-ME OPTIONAL NAME IN CONCOURSE UI",
				Script: "./gradlew CHANGE-ME",
				Docker: Docker{
					Image: "CHANGE-ME:tag",
				},
			},
		},
	}

	out, err := yaml.Marshal(manifest)
	if err != nil {
		return
	}

	err = s.fs.WriteFile(".halfpipe.io", out, 0777)
	return
}
