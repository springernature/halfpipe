package actions

import (
	"fmt"
	"github.com/springernature/halfpipe/manifest"
)

func (a *Actions) updateSteps(task manifest.Update, man manifest.Manifest) Steps {
	update := Step{
		Name: "Sync workflow with halfpipe manifest",
		ID:   "sync",
		Run:  "halfpipe-update-workflow",
		Env: Env{
			"HALFPIPE_FILE_PATH": a.halfpipeFilePath,
		},
	}

	push := Step{
		Name: "Commit and push changes to workflow",
		If:   "steps.sync.outputs.synced == 'false'",
		Run: `git config user.name halfpipe-io
git config user.email halfpipe-io@springernature.com
if git commit -am "[halfpipe] synced workflow $GITHUB_WORKFLOW with halfpipe manifest" && git push; then
  echo ':white_check_mark: Halfpipe successfully updated the workflow' >> $GITHUB_STEP_SUMMARY
  echo >> $GITHUB_STEP_SUMMARY
  echo 'This happened because the workflow was generated from a halfpipe manifest with the ` + "`update-pipeline`" + ` feature enabled. It keeps the workflow in sync with the halfpipe manifest.' >> $GITHUB_STEP_SUMMARY
  echo >> $GITHUB_STEP_SUMMARY
  echo '[Halfpipe Documentation](https://ee.public.springernature.app/rel-eng/halfpipe/features/#update_pipeline)' >> $GITHUB_STEP_SUMMARY
else
  echo ':x: Halfpipe failed to update the workflow' >> $GITHUB_STEP_SUMMARY
  echo >> $GITHUB_STEP_SUMMARY
  echo 'This may have happened because newer git commits have already been pushed. Check for newer pipeline runs or manually trigger the workflow.' >> $GITHUB_STEP_SUMMARY
  echo >> $GITHUB_STEP_SUMMARY
  echo '[Halfpipe Documentation](https://ee.public.springernature.app/rel-eng/halfpipe/features/#update_pipeline)' >> $GITHUB_STEP_SUMMARY
  exit 1
fi
`,
	}

	steps := Steps{update, push}

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
