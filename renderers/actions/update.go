package actions

import (
	"fmt"
	"github.com/springernature/halfpipe/manifest"
)

func (a *Actions) updateSteps(task manifest.Update, man manifest.Manifest) Steps {
	steps := Steps{}

	cdPrefix := ""
	if a.workingDir != "" {
		cdPrefix = fmt.Sprintf("cd %s; ", a.workingDir)
	}

	update := Step{
		Name: "Sync workflow with halfpipe manifest",
		ID:   "sync",
		Uses: "docker://eu.gcr.io/halfpipe-io/halfpipe-auto-update",
		With: With{
			"args":       fmt.Sprintf(`-c "%supdate-actions-workflow"`, cdPrefix),
			"entrypoint": "/bin/bash",
		},
		Env: Env{
			"HALFPIPE_FILE_PATH": a.halfpipeFilePath,
		},
	}

	push := Step{
		Name: "Commit and push changes to workflow",
		If:   "steps.sync.outputs.synced == 'false'",
		Run: `git config user.name halfpipe-io
git config user.email halfpipe-io@springernature.com
git commit -am "[halfpipe] synced workflow $GITHUB_WORKFLOW with halfpipe manifest"
git push
echo 'Halfpipe successfully updated the workflow' >> $GITHUB_STEP_SUMMARY
echo >> $GITHUB_STEP_SUMMARY
echo 'This happened because the workflow was generated from a halfpipe manifest with the ` + "`update-pipeline`" + ` feature enabled. It keeps the workflow in sync with the halfpipe manifest.' >> $GITHUB_STEP_SUMMARY
echo >> $GITHUB_STEP_SUMMARY
echo '[Halfpipe Documentation](https://ee.public.springernature.app/rel-eng/halfpipe/features/#update_pipeline)' >> $GITHUB_STEP_SUMMARY
`,
	}

	if man.FeatureToggles.UpdateActions() {
		steps = Steps{update, push}
	}

	if task.TagRepo {
		tag := man.PipelineName() + "/v$BUILD_VERSION"
		tagStep := Step{
			Name: "Tag commit with " + tag,
			Run:  fmt.Sprintf("git tag -f %s\ngit push origin %s", tag, tag),
		}
		steps = append(steps, tagStep)
	}

	return steps
}
