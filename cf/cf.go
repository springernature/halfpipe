package cf

import (
	"code.cloudfoundry.org/cli/util/manifest"
	"github.com/cloudfoundry/bosh-cli/director/template"
)

type ManifestReader func(pathToManifest string, pathsToVarsFiles []string, vars []template.VarKV) ([]manifest.Application, error)
