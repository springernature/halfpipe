package actions

import (
	"strings"
)

func (a *Actions) saveArtifacts(paths []string) []Step {
	return saveArtifactSteps(paths, "artifacts")
}

func (a *Actions) saveArtifactsOnFailure(paths []string) []Step {
	steps := saveArtifactSteps(paths, "artifacts-failure")
	steps[0].If = "failure()"
	steps[1].If = "failure()"
	return steps
}

func saveArtifactSteps(paths []string, name string) []Step {
	return []Step{
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

func (a *Actions) restoreArtifacts() []Step {
	return []Step{
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
