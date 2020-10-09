package actions

import "gopkg.in/yaml.v2"

type Push struct {
	Branches []string `json:"branches,omitempty" yaml:"branches,omitempty"`
	Paths    []string `json:"paths,omitempty" yaml:"paths,omitempty"`
}

type On struct {
	Push Push `json:"push,omitempty" yaml:"push,omitempty"`
}

type Container struct {
	Image string `json:"image,omitempty" yaml:"image,omitempty"`
}

type Step struct {
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	Uses string `json:"uses,omitempty" yaml:"uses,omitempty"`
	Run  string `json:"run,omitempty" yaml:"run,omitempty"`
}

type Job struct {
	Name      string    `json:"name,omitempty" yaml:"name,omitempty"`
	RunsOn    string    `json:"runs-on,omitempty" yaml:"runs-on,omitempty"`
	Container Container `json:"container,omitempty" yaml:"container,omitempty"`
	Steps     []Step    `json:"steps,omitempty" yaml:"steps,omitempty"`
}

type Actions struct {
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	On On `json:"on,omitempty" yaml:"on,omitempty"`

	Jobs yaml.MapSlice `json:"jobs,omitempty" yaml:"jobs,omitempty"`
}
