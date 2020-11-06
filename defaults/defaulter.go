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

type DefaultValuesDefaulter interface {
	Apply(defaults Defaults) (updated manifest.DefaultValues)
}

type CFOnPrem struct {
	Username string
	Password string
}

type CFSnPaaS struct {
	Username string
	Password string
	Org      string
	API      string
}

type CFDefaults struct {
	ManifestPath string
	OnPrem       CFOnPrem
	SnPaaS       CFSnPaaS
	TestDomains  map[string]string
	Version      string
}

type DockerDefaults struct {
	Username       string
	Password       string
	FilePath       string
	ComposeFile    string
	ComposeService string
}

type ArtifactoryDefaults struct {
	Username string
	Password string
	URL      string
}

type ConcourseDefaults struct {
	URL      string
	Username string
	Password string
}

type MarkLogicDefaults struct {
	Username string
	Password string
}

type Aux struct {
	RepoPrivateKey  string
	RepoAccessToken string
	Timeout         string
	SlackToken      string
}

type Defaults struct {
	Project project.Data

	Aux         Aux
	CF          CFDefaults
	Docker      DockerDefaults
	Artifactory ArtifactoryDefaults
	Concourse   ConcourseDefaults
	MarkLogic   MarkLogicDefaults

	triggersDefaulter      TriggersDefaulter
	tasksDefaulter         TasksDefaulter
	defaultValuesDefaulter DefaultValuesDefaulter
}

func (d Defaults) Apply(original manifest.Manifest) (updated manifest.Manifest) {
	updated = original

	if updated.FeatureToggles.UpdatePipeline() {
		updated.Tasks = append(manifest.TaskList{manifest.Update{}}, updated.Tasks...)
	}

	updated.Triggers = d.triggersDefaulter.Apply(updated.Triggers, d, original)
	updated.Tasks = d.tasksDefaulter.Apply(updated.Tasks, d, updated)
	updated.DefaultValues = d.defaultValuesDefaulter.Apply(d)
	return updated
}

func New(defaultvValues Defaults, project project.Data) Defaults {
	defaultvValues.Project = project
	defaultvValues.triggersDefaulter = NewTriggersDefaulter()
	defaultvValues.tasksDefaulter = NewTaskDefaulter()
	defaultvValues.defaultValuesDefaulter = NewDefaultValuesDefaulter()

	return defaultvValues
}
