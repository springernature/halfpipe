package cf

import (
	"code.cloudfoundry.org/cli/util/manifestparser"
	"github.com/cloudfoundry/bosh-cli/director/template"
)

type ManifestReader func(pathToManifest string, pathsToVarsFiles []string, vars []template.VarKV) (manifestparser.Manifest, error)

func Routes(app manifestparser.Application) (rs []string) {
	rawRoutes := []any{}

	if app.RemainingManifestFields["routes"] != nil {
		rawRoutes = app.RemainingManifestFields["routes"].([]any)
	}

	for _, r := range rawRoutes {
		route := r.(map[any]any)["route"].(string)
		rs = append(rs, route)
	}
	return rs
}

func Buildpacks(app manifestparser.Application) (bps []string) {
	raw := []any{}

	if app.RemainingManifestFields["buildpacks"] != nil {
		raw = app.RemainingManifestFields["buildpacks"].([]any)
	}

	for _, r := range raw {
		bps = append(bps, r.(string))
	}
	return bps
}
