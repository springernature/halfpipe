package actions

type Push struct {
	Branches []string `json:"branches,omitempty" yaml:"branches,omitempty"`
	Paths    []string `json:"paths,omitempty" yaml:"paths,omitempty"`
}

type On struct {
	Push Push `json:"push,omitempty" yaml:"push,omitempty"`
}

type Actions struct {
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	On On `json:"on,omitempty" yaml:"on,omitempty"`
}
