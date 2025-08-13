package actions

import (
	"path/filepath"
	"strings"
)

func (a *Actions) saveArtifacts(paths []string) Steps {
	return a.saveArtifactSteps(paths, "artifacts")
}

func (a *Actions) saveArtifactsOnFailure(paths []string) Steps {
	steps := a.saveArtifactSteps(paths, "artifacts-failure")
	steps[0].If = "failure()"
	steps[1].If = "failure()"
	return steps
}

func (a *Actions) saveArtifactSteps(paths []string, name string) Steps {
	// in halfpipe manifest the paths are specified relative to the halfpipe file
	// we need to convert the paths to be relative to github workspace
	for i, path := range paths {
		paths[i] = filepath.Clean(filepath.Join(a.workingDir, path))
	}

	return Steps{
		{
			Name:             "Package " + name,
			Run:              "tar -cvf /tmp/halfpipe-artifacts-${{ github.job }}.tar " + strings.Join(paths, " "),
			WorkingDirectory: "${{ github.workspace }}",
		},
		{
			Name: "Upload " + name,
			Uses: "actions/upload-artifact@ea165f8d65b6e75b540449e92b4886f43607fa02",
			With: With{
				"name":           name + "-${{ github.job }}",
				"path":           "/tmp/halfpipe-artifacts-${{ github.job }}.tar",
				"retention-days": 2,
			},
		},
	}
}

func (a *Actions) restoreArtifacts() Steps {
	return Steps{
		{
			Name: "Download artifacts",
			Uses: "actions/download-artifact@634f93cb2916e3fdff6788551b99b062d0335ce0",
			With: With{
				"merge-multiple": true,
			},
		},
		{
			Name:             "Extract artifacts",
			Run:              "for f in halfpipe-artifacts-*.tar; do tar -xvf $f; done; rm halfpipe-artifacts-*.tar",
			WorkingDirectory: "${{ github.workspace }}",
		},
	}
}
