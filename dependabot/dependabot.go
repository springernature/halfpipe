package dependabot

import "github.com/sirupsen/logrus"

type Dependabot interface {
	Resolve() (Config, error)
}

type dependabot struct {
	config DependabotConfig
	walker Walker
	filter Filter
	render Render
}

func (d dependabot) Resolve() (c Config, err error) {
	logrus.Debug("Walking the filesystem")
	files, err := d.walker.Walk(d.config.Depth, d.config.SkipFolders)
	if err != nil {
		return
	}

	logrus.Debugf("Found '%d' files", len(files))
	logrus.Debug("Filtering files")
	filtered := d.filter.Filter(files, d.config.SkipEcosystem)
	logrus.Debugf("Filtered out '%d' files", len(filtered))
	for _, filteredFile := range filtered {
		logrus.Debugf("'%s'", filteredFile)
	}
	logrus.Debug("Filtering files")

	c = d.render.Render(filtered)
	return
}

func New(config DependabotConfig, walker Walker, filter Filter, render Render) Dependabot {
	return dependabot{
		config: config,
		walker: walker,
		filter: filter,
		render: render,
	}
}
