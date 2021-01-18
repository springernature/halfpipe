package actions

import (
	"path"
	"sort"
	"strings"
)

func (a *Actions) saveArtifacts(paths []string) Step {
	for i, v := range paths {
		paths[i] = path.Join(a.workingDir, v)
		a.savedArtifacts[paths[i]] = true
	}
	return Step{
		Name: "Save artifacts",
		Uses: "actions/upload-artifact@v2",
		With: With{
			{"name", "artifacts"},
			{"path", strings.Join(paths, "\n") + "\n"},
		},
	}
}

func (a *Actions) saveArtifactsOnFailure(paths []string) Step {
	step := a.saveArtifacts(paths)
	step.Name += " (failure)"
	step.If = "failure()"
	return step
}

func (a *Actions) restoreArtifacts() Step {
	paths := []string{}
	for k := range a.savedArtifacts {
		paths = append(paths, k)
	}
	sort.Strings(paths)
	return Step{
		Name: "Restore artifacts",
		Uses: "actions/download-artifact@v2",
		With: With{
			{"name", "artifacts"},
			{"path", strings.Join(paths, "\n") + "\n"},
		},
	}
}
