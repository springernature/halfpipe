package samplegenerator

import (
	"errors"
	"fmt"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
	"github.com/springernature/halfpipe/project"
	"gopkg.in/yaml.v2"
	"strings"
)

var ErrHalfpipeAlreadyExists = errors.New("'.halfpipe.io.yml' already exists")

type SampleGenerator interface {
	Generate() (err error)
}

type sampleGenerator struct {
	fs              afero.Afero
	projectResolver project.Project
	currentDir      string
}

func NewSampleGenerator(fs afero.Afero, projectResolver project.Project, currentDir string) SampleGenerator {
	return sampleGenerator{
		fs:              fs,
		projectResolver: projectResolver,
		currentDir:      currentDir,
	}
}

func (s sampleGenerator) Generate() (err error) {
	exists, err := s.fs.Exists(".halfpipe.io.yml")
	if err != nil {
		return err
	}

	if exists {
		return ErrHalfpipeAlreadyExists
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

	proj, err := s.projectResolver.Parse(s.currentDir, true, config.HalfpipeFilenameOptions)
	if err != nil {
		return err
	}

	if proj.BasePath != "" {
		man.Triggers = append(man.Triggers, manifest.GitTrigger{
			WatchedPaths: []string{proj.BasePath},
		})
	}
	man.Pipeline = createPipelineName(proj)

	man.FeatureToggles = manifest.FeatureToggles{manifest.FeatureUpdatePipeline}

	out, err := yaml.Marshal(man)
	if err != nil {
		return err
	}

	return s.fs.WriteFile(".halfpipe.io.yml", out, 0644)
}

func createPipelineName(project project.Data) string {
	if project.BasePath == "" {
		return project.RootName
	}
	return strings.Replace(fmt.Sprintf("%s-%s", project.RootName, project.BasePath), "/", "-", -1)
}
