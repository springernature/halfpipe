package defaults

import (
	"github.com/springernature/halfpipe/manifest"
	"github.com/springernature/halfpipe/project"
)

type DefaulterNew interface {
	Apply(original manifest.Manifest) (updated manifest.Manifest)
}

type TriggersDefaulter interface {
	Apply(original manifest.TriggerList, defaults DefaultsNew, man manifest.Manifest) (updated manifest.TriggerList)
}

type TasksRenamer interface {
	Apply(original manifest.TaskList) (updated manifest.TaskList)
}

type TasksDefaulter interface {
	Apply(original manifest.TaskList, defaults DefaultsNew) (updated manifest.TaskList)
}

type DefaultsNew struct {
	RepoPrivateKey       string
	CfUsername           string
	CfPassword           string
	CfUsernameSnPaas     string
	CfPasswordSnPaas     string
	CfOrgSnPaas          string
	CfAPISnPaas          string
	CfManifest           string
	CfTestDomains        map[string]string
	DockerUsername       string
	DockerPassword       string
	Project              project.Data
	DockerComposeService string
	ArtifactoryUsername  string
	ArtifactoryPassword  string
	ArtifactoryURL       string
	ConcourseURL         string
	ConcourseUsername    string
	ConcoursePassword    string
	Timeout              string

	triggersDefaulter TriggersDefaulter
	tasksRenamer      TasksRenamer
	tasksDefaulter    TasksDefaulter
}

func (d DefaultsNew) Apply(original manifest.Manifest) (updated manifest.Manifest) {
	updated = original

	if updated.FeatureToggles.UpdatePipeline() {
		updated.Tasks = append(manifest.TaskList{manifest.Update{}}, updated.Tasks...)
	}

	updated.Triggers = d.triggersDefaulter.Apply(updated.Triggers, d, original)
	updated.Tasks = d.tasksRenamer.Apply(updated.Tasks)
	updated.Tasks = d.tasksDefaulter.Apply(updated.Tasks, d)

	return
}

func NewNewDefaulter(project project.Data) DefaultsNew {
	d := DefaultValuesNew
	d.Project = project
	d.triggersDefaulter = NewTriggersDefaulter()
	d.tasksRenamer = NewTasksRenamer()
	d.tasksDefaulter = NewTaskDefaulter()
	return d
}
