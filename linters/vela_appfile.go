package linters

import (
	"gopkg.in/yaml.v3"
)

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
