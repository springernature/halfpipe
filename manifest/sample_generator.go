package manifest

import (
	"errors"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
	"github.com/springernature/halfpipe/project"
	"fmt"
	"strings"
)

var ErrHalfpipeAlreadyExists = errors.New("'.halfpipe.io' already exists")

type SampleGenerator interface {
	Generate() (err error)
}

type sampleGenerator struct {
	fs              afero.Afero
	projectResolver project.ProjectResolver
	currentDir      string
}

func NewSampleGenerator(fs afero.Afero, projectResolver project.ProjectResolver, currentDir string) SampleGenerator {
	return sampleGenerator{
		fs:              fs,
		projectResolver: projectResolver,
		currentDir:      currentDir,
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

	project, err := s.projectResolver.Parse(s.currentDir)
	if err != nil {
		return
	}

	if project.BasePath != "" {
		manifest.Repo.WatchedPaths = append(manifest.Repo.WatchedPaths, project.BasePath)
	}
	manifest.Pipeline = createPipelineName(project)

	out, err := yaml.Marshal(manifest)
	if err != nil {
		return
	}

	err = s.fs.WriteFile(".halfpipe.io", out, 0777)
	return
}

func createPipelineName(project project.Project) string {
	if project.BasePath == "" {
		return project.RootName
	}
	return strings.Replace(fmt.Sprintf("%s-%s", project.RootName, project.BasePath), "/", "-", -1)
}