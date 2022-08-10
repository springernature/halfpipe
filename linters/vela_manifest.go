package linters

//
//import (
//	"fmt"
//	"github.com/springernature/halfpipe/linters/result"
//	"github.com/springernature/halfpipe/manifest"
//	"gopkg.in/yaml.v3"
//
//	"io/fs"
//)
//
//type VelaManifestLinter struct {
//	fs fs.FS
//}
//
//func (v VelaManifestLinter) Lint(man manifest.Manifest) (lr result.LintResult) {
//
//	var deployKateeTasks []manifest.DeployKatee
//	for _, task := range man.Tasks {
//		switch t := task.(type) {
//		case manifest.DeployKatee:
//			deployKateeTasks = append(deployKateeTasks, t)
//		}
//	}
//
//	for _, kateeTask := range deployKateeTasks {
//		velaAppFile, err := fs.ReadFile(v.fs, kateeTask.VelaManifest)
//		velaManifest := unMarshallVelaManifest(velaAppFile)
//
//		for _, comp := range velaManifest.Spec.Components {
//			for _, e := range comp.Properties.Env {
//				strings.Hase.Value
//			}
//		}
//
//		if err != nil {
//			lr.Errors = append(lr.Errors, err)
//		}
//	}
//
//	return lr
//}
//
//type VelaManifest struct {
//	Kind string     `yaml:"kind"`
//	Spec Components `yaml:"spec"`
//}
//
//type Components struct {
//	Components []Component `yaml:"components"`
//}
//
//type Properties struct {
//	Image string `yaml:"image"`
//	Env   []Env  `yaml:"env"`
//}
//
//type Component struct {
//	Name       string     `yaml:"name"`
//	Type       string     `yaml:"type"`
//	Properties Properties `yaml:"properties"`
//}
//
//type Env struct {
//	Name  string `yaml:"name"`
//	Value string `yaml:"value"`
//}
//
//func unMarshallVelaManifest(bytes []byte) VelaManifest {
//	var vm VelaManifest
//	err := yaml.Unmarshal(bytes, &vm)
//	if err != nil {
//		fmt.Println(err)
//	}
//	fmt.Printf("%+v\n", vm)
//	return vm
//}
