package dependabot

import "github.com/sirupsen/logrus"

type Dependabot interface {
	Resolve() (Config, error)
}

type dependabot struct {
	walker Walker
	filter Filter
	render Renderer
}

func (d dependabot) Resolve() (c Config, err error) {
	logrus.Debug("Walking the filesystem")
	files, err := d.walker.Walk()
	if err != nil {
		return
	}

	logrus.Debugf("Found '%d' files", len(files))
	logrus.Debug("Filtering files")
	filtered := d.filter.Filter(files)
	logrus.Debugf("Filtered out '%d' files", len(filtered))
	for path, filteredFile := range filtered {
		logrus.Debugf("'%s' -> '%s'", path, filteredFile)
	}
	logrus.Debug("Filtering files")

	c = d.render(filtered)
	return
}

func New(walker Walker, filter Filter, render Renderer) Dependabot {
	return dependabot{
		walker: walker,
		filter: filter,
		render: render,
	}
}
