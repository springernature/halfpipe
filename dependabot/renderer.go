package dependabot

import (
	"fmt"
	"path/filepath"
)

type Render interface {
	Render(paths []string) Config
}

type render struct {
}

func (r render) renderPath(path string) (d Dependency) {
	fileName := filepath.Base(path)
	dir := filepath.Dir(path)
	if dir == "." {
		dir = "/"
	} else {
		dir = fmt.Sprintf("/%s", dir)
	}

	if ecosystem, ok := SupportedFiles[fileName]; ok {
		d.PackageEcosystem = ecosystem
		d.Directory = dir
		d.Schedule.Interval = "daily"
	}

	return
}

func (r render) renderPaths(paths []string) (d []Dependency) {
	for _, path := range paths {
		d = append(d, r.renderPath(path))
	}
	return
}

func (r render) Render(paths []string) Config {
	return Config{
		Version: 2,
		Updates: r.renderPaths(paths),
	}
}

func NewRender() Render {
	return render{}
}
