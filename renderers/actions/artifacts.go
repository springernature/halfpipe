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
			Uses: "actions/upload-artifact@65c4c4a1ddee5b72f698fdd19549f0f0fb45cf08",
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
			Uses: "actions/download-artifact@fa0a91b85d4f404e444e00e005971372dc801d16",
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
