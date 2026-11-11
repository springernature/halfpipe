package concourse

import (
	"fmt"
	"path"
	"strings"

	"github.com/concourse/concourse/atc"
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/manifest"
)

func (c Concourse) deployKateeJob(task manifest.DeployKatee, man manifest.Manifest, basePath string) (job atc.JobConfig) {
	job.Name = task.GetName()
	job.Serial = true

	velaManifest := path.Join(gitDir, basePath, task.VelaManifest)
	if strings.HasPrefix(task.VelaManifest, fmt.Sprintf("../%s/", artifactsInDir)) {
		velaManifest = strings.TrimPrefix(task.VelaManifest, "../")
	}

	params := atc.Params{
		"credentials": config.VaultSecrets.KateeKey(man.Team, task.Namespace, task.Env),
		"velafile":    velaManifest,
		"gitRefPath":  "git/.git/ref",
		"vars":        task.Vars,
		"eaid":        man.OpsLevel.System,
	}
	// Like run tasks, which export BUILD_VERSION when the pipeline has a version.
	if man.FeatureToggles.UpdatePipeline() {
		params["buildVersionPath"] = path.Join(versionName, "version")
	}

	step := stepWithAttemptsAndTimeout(&atc.PutStep{
		Name:     "Deploy to Katee",
		Resource: task.Namespace,
		Params:   params,
		NoGet:    true,
	}, task.TaskBase)

	job.PlanSequence = []atc.Step{
		step,
	}

	return
}
