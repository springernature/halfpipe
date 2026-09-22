package defaults

import (
	"github.com/springernature/halfpipe/manifest"
)

func uploadSLOsDefaulter(original manifest.UploadSLOs) (updated manifest.UploadSLOs) {
	updated = original

	if original.Folder == "" {
		updated.Folder = "slos"
	}

	return updated
}
