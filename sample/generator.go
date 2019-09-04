package sample

import (
	"errors"
	"github.com/springernature/halfpipe/manifest"

	"fmt"
	"strings"

	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/project"
	"gopkg.in/yaml.v2"
)

var ErrHalfpipeAlreadyExists = errors.New("'.halfpipe.io' already exists")

type Generator interface {
	Generate() (err error)
}

type sampleGenerator struct {
	fs              afero.Afero
	projectResolver project.Project
	currentDir      string
}

func NewSampleGenerator(fs afero.Afero, projectResolver project.Project, currentDir string) Generator {
	return sampleGenerator{
		fs:              fs,
		projectResolver: projectResolver,
		currentDir:      currentDir,
	}
}

func (s sampleGenerator) Generate() (err error) {
	proj, err := s.projectResolver.Parse(s.currentDir)
	if err != nil {
		return
	}

	if proj.HalfpipeFilePath != "" {
		exists, e := s.fs.Exists(proj.HalfpipeFilePath)
		if e != nil {
			err = e
			return
		}

		if exists {
			err = ErrHalfpipeAlreadyExists
			return
		}
	}

	man := manifest.Manifest{
		Team: "CHANGE-ME",

		Tasks: []manifest.Task{
			manifest.Run{
				Type:   "run",
				Name:   "CHANGE-ME OPTIONAL NAME IN CONCOURSE UI",
				Script: "./gradlew CHANGE-ME",
				Docker: manifest.Docker{
					Image: "CHANGE-ME:tag",
				},
			},
		},
	}

	if proj.BasePath != "" {
		man.Triggers = append(man.Triggers, manifest.GitTrigger{
			Type:         "git",
			WatchedPaths: []string{proj.BasePath},
		})
	}
	man.Pipeline = createPipelineName(proj)

	man.FeatureToggles = manifest.FeatureToggles{manifest.FeatureUpdatePipeline}

	out, err := yaml.Marshal(man)
	if err != nil {
		return
	}

	err = s.fs.WriteFile(".halfpipe.io", out, 0777)
	return
}

func createPipelineName(project project.Data) string {
	if project.BasePath == "" {
		return project.RootName
	}
	return strings.Replace(fmt.Sprintf("%s-%s", project.RootName, project.BasePath), "/", "-", -1)
}
