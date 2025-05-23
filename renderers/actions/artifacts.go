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
			Run:              "tar -cvf /tmp/halfpipe-artifacts.tar " + strings.Join(paths, " "),
			WorkingDirectory: "${{ github.workspace }}",
		},
		{
			Name: "Upload " + name,
			Uses: "actions/upload-artifact@ea165f8d65b6e75b540449e92b4886f43607fa02",
			With: With{
				"name":           name,
				"path":           "/tmp/halfpipe-artifacts.tar",
				"retention-days": 2,
			},
		},
	}
}

func (a *Actions) restoreArtifacts() Steps {
	return Steps{
		{
			Name: "Download artifacts",
			Uses: "actions/download-artifact@d3f86a106a0bac45b974a628896c90dbdf5c8093",
			With: With{
				"name": "artifacts",
			},
		},
		{
			Name:             "Extract artifacts",
			Run:              "tar -xvf halfpipe-artifacts.tar; rm halfpipe-artifacts.tar",
			WorkingDirectory: "${{ github.workspace }}",
		},
	}
}
