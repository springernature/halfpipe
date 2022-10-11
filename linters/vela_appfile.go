package linters

import (
	"github.com/spf13/afero"
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

func (v VelaManifestLinter) Lint(man manifest.Manifest) (lr LintResult) {
	var deployKateeTasks []manifest.DeployKatee
	for _, task := range man.Tasks {
		switch t := task.(type) {
		case manifest.DeployKatee:
			deployKateeTasks = append(deployKateeTasks, t)
		}
	}

	for _, kateeTask := range deployKateeTasks {
		err := CheckFile(v.Fs, kateeTask.VelaManifest, false)
		if err != nil {
			lr.Add(err)
			return
		}

		velaAppFile, err := ReadFile(v.Fs, kateeTask.VelaManifest)
		if err != nil {
			lr.Add(err)
		}

		velaManifest, e := unMarshallVelaManifest([]byte(velaAppFile))
		if e != nil {
			lr.Add(ErrFileInvalid.WithValue(e.Error()))
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
							lr.Add(ErrVelaVariableMissing.WithValue(secretName).WithFile(kateeTask.VelaManifest))
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
