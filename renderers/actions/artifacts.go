package actions

import (
	"strings"
)

func (a *Actions) saveArtifacts(paths []string) Steps {
	return saveArtifactSteps(paths, "artifacts")
}

func (a *Actions) saveArtifactsOnFailure(paths []string) Steps {
	steps := saveArtifactSteps(paths, "artifacts-failure")
	steps[0].If = "failure()"
	steps[1].If = "failure()"
	return steps
}

func saveArtifactSteps(paths []string, name string) Steps {
	return Steps{
		{
			Name: "Package " + name,
			Run:  "tar -cvf /tmp/halfpipe-artifacts.tar " + strings.Join(paths, " "),
		},
		{
			Name: "Upload " + name,
			Uses: "actions/upload-artifact@v2",
			With: With{
				{"name", name},
				{"path", "/tmp/halfpipe-artifacts.tar"},
			},
		},
	}
}

func (a *Actions) restoreArtifacts() Steps {
	return Steps{
		{
			Name: "Download artifacts",
			Uses: "actions/download-artifact@v2",
			With: With{
				{"name", "artifacts"},
				{"path", a.workingDir},
			},
		},
		{
			Name: "Extract artifacts",
			Run:  "tar -xvf halfpipe-artifacts.tar; rm halfpipe-artifacts.tar",
		},
	}
}
