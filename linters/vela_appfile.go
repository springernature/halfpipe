package linters

import (
	"errors"
	"fmt"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/linters/filechecker"
	"github.com/springernature/halfpipe/linters/linterrors"
	"github.com/springernature/halfpipe/linters/result"
	"github.com/springernature/halfpipe/manifest"
	"gopkg.in/yaml.v3"
	"strings"
)

type VelaManifestLinter struct {
	Fs afero.Afero
}

func NewVelaManifestLinter(fs afero.Afero) VelaManifestLinter {
	return VelaManifestLinter{fs}
}

func (v VelaManifestLinter) Lint(man manifest.Manifest) (lr result.LintResult) {
	var deployKateeTasks []manifest.DeployKatee
	for _, task := range man.Tasks {
		switch t := task.(type) {
		case manifest.DeployKatee:
			deployKateeTasks = append(deployKateeTasks, t)
		}
	}

	for _, kateeTask := range deployKateeTasks {
		err := filechecker.CheckFile(v.Fs, kateeTask.VelaManifest, false)
		if err != nil {
			lr.AddError(err)
			return
		}

		velaAppFile, err := v.Fs.ReadFile(kateeTask.VelaManifest)
		if err != nil {
			lr.AddError(linterrors.NewFileError(kateeTask.VelaManifest, "does not exist"))
		}

		velaManifest, e := unMarshallVelaManifest(velaAppFile)
		if e != nil {
			lr.AddError(errors.New("vela manifest is invalid"))
			return
		}

		for _, com := range velaManifest.Spec.Components {
			for _, sec := range com.Properties.Env {
				if strings.HasPrefix(sec.Value, "${") {
					secretName := strings.ReplaceAll(sec.Value, "${", "")
					secretName = strings.ReplaceAll(secretName, "}", "")

					vars := kateeTask.Vars
					if _, ok := vars[secretName]; !ok {
						if secretName != "BUILD_VERSION" && secretName != "GIT_REVISION" {
							lr.AddError(fmt.Errorf("vela manifest variable %s is not specified in halfpipe manifest", secretName))
						}
					}
				}
			}
		}
	}
	return lr
}

type VelaManifest struct {
	Kind string     `yaml:"kind"`
	Spec Components `yaml:"spec"`
}

type Components struct {
	Components []Component `yaml:"components"`
}

type Properties struct {
	Image string `yaml:"image"`
	Env   []Env  `yaml:"env"`
}

type Component struct {
	Name       string     `yaml:"name"`
	Type       string     `yaml:"type"`
	Properties Properties `yaml:"properties"`
}

type Env struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

func unMarshallVelaManifest(bytes []byte) (vm VelaManifest, e error) {
	e = yaml.Unmarshal(bytes, &vm)
	if e != nil {
		return vm, e
	}
	return vm, nil
}
