package defaults

import (
	"github.com/springernature/halfpipe/manifest"
	"github.com/springernature/halfpipe/project"
)

type TriggersDefaulter interface {
	Apply(original manifest.TriggerList, defaults Defaults, man manifest.Manifest) (updated manifest.TriggerList)
}

type TasksDefaulter interface {
	Apply(original manifest.TaskList, defaults Defaults, man manifest.Manifest) (updated manifest.TaskList)
}

type BuildHistoryDefaulter interface {
	Apply(original manifest.TaskList, defaults Defaults) (updated manifest.TaskList)
}

type DefaultsI interface {
	Apply(original manifest.Manifest) (updated manifest.Manifest)
}

type Defaults struct {
	RepoPrivateKey string

	CfUsername       string
	CfPassword       string
	CfUsernameSnPaas string
	CfPasswordSnPaas string
	CfOrgSnPaas      string
	CfAPISnPaas      string
	CfManifest       string
	CfTestDomains    map[string]string
	CfCliVersion     string

	DockerUsername string
	DockerPassword string
	DockerfilePath string

	Project project.Data

	DockerComposeFile    string
	DockerComposeService string

	ArtifactoryUsername string
	ArtifactoryPassword string
	ArtifactoryURL      string

	ConcourseURL      string
	ConcourseUsername string
	ConcoursePassword string

	MarkLogicUsername string
	MarkLogicPassword string

	Timeout string

	triggersDefaulter TriggersDefaulter
	tasksDefaulter    TasksDefaulter
}

func (d Defaults) Apply(original manifest.Manifest) (updated manifest.Manifest) {
	updated = original

	if updated.FeatureToggles.UpdatePipeline() {
		updated.Tasks = append(manifest.TaskList{manifest.Update{}}, updated.Tasks...)
	}

	updated.Triggers = d.triggersDefaulter.Apply(updated.Triggers, d, original)
	updated.Tasks = d.tasksDefaulter.Apply(updated.Tasks, d, updated)

	return updated
}

func New(project project.Data) Defaults {
	d := DefaultValues
	d.Project = project

	d.triggersDefaulter = NewTriggersDefaulter()
	d.tasksDefaulter = NewTaskDefaulter()

	return d
}
