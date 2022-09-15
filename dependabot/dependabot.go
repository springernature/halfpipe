package dependabot

type Dependabot interface {
	Resolve() (Config, error)
}

type dependabot struct {
	config DependabotConfig
	walker Walker
}

func (d dependabot) Resolve() (Config, error) {
	_, err := d.walker.Walk(d.config.Depth, d.config.SkipFolders)
	return Config{}, err
}

func New(config DependabotConfig, walker Walker) Dependabot {
	return dependabot{
		config: config,
		walker: walker,
	}
}
