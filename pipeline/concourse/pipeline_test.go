package concourse

import (
	cfManifest "code.cloudfoundry.org/cli/util/manifest"

	"github.com/cloudfoundry/bosh-cli/director/template"
	"github.com/spf13/afero"
)

func testPipeline() renderer {
	cfManifestReader := func(pathToManifest string, pathsToVarsFiles []string, vars []template.VarKV) ([]cfManifest.Application, error) {
		return []cfManifest.Application{
			{
				Name:   "test-name",
				Routes: []string{"test-route"},
			},
		}, nil
	}

	return NewRenderer(cfManifestReader, afero.Afero{Fs: afero.NewMemMapFs()})
}

/*
  ___     ___    _  _   _____     ___   _   _   _____     ___   _____   _   _   ___   ___     _  _   ___   ___   ___
 |   \   / _ \  | \| | |_   _|   | _ \ | | | | |_   _|   / __| |_   _| | | | | | __| | __|   | || | | __| | _ \ | __|
 | |) | | (_) | | .` |   | |     |  _/ | |_| |   | |     \__ \   | |   | |_| | | _|  | _|    | __ | | _|  |   / | _|
 |___/   \___/  |_|\_|   |_|     |_|    \___/    |_|     |___/   |_|    \___/  |_|   |_|     |_||_| |___| |_|_\ |___|

  _   _   ___   ___     _____   _  _   ___      ___    _____   _  _   ___   ___     _____   ___   ___   _____     ___   ___   _      ___   ___
 | | | | / __| | __|   |_   _| | || | | __|    / _ \  |_   _| | || | | __| | _ \   |_   _| | __| / __| |_   _|   | __| |_ _| | |    | __| / __|
 | |_| | \__ \ | _|      | |   | __ | | _|    | (_) |   | |   | __ | | _|  |   /     | |   | _|  \__ \   | |     | _|   | |  | |__  | _|  \__ \
  \___/  |___/ |___|     |_|   |_||_| |___|    \___/    |_|   |_||_| |___| |_|_\     |_|   |___| |___/   |_|     |_|   |___| |____| |___| |___/

*/
