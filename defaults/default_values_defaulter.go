package defaults

import "github.com/springernature/halfpipe/manifest"

type defaultValues struct {
}

func NewDefaultValuesDefaulter() DefaultValuesDefaulter {
	return defaultValues{}
}

func (t defaultValues) Apply(defaults Defaults) manifest.DefaultValues {
	return manifest.DefaultValues{
		SlackToken:      defaults.Aux.SlackToken,
		RepoAccessToken: defaults.Aux.RepoAccessToken,
	}
}
