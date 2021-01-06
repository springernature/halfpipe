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

type Defaults struct {
	Project project.Data

	RepoPrivateKey string
	CF             CFDefaults
	Docker         DockerDefaults
	Artifactory    ArtifactoryDefaults
	Concourse      ConcourseDefaults
	MarkLogic      MarkLogicDefaults

	Timeout string

	triggersDefaulter TriggersDefaulter
	tasksDefaulter    TasksDefaulter

	RemoveUpdatePipelineFeature bool
}

func (d Defaults) Apply(original manifest.Manifest) (updated manifest.Manifest) {
	updated = original

	if d.RemoveUpdatePipelineFeature {
		updated.FeatureToggles = []string{}
		for _, f := range original.FeatureToggles {
			if f != manifest.FeatureUpdatePipeline {
				updated.FeatureToggles = append(updated.FeatureToggles, f)
			}
		}
	}

	updated.Triggers = d.triggersDefaulter.Apply(updated.Triggers, d, original)
	updated.Tasks = d.tasksDefaulter.Apply(updated.Tasks, d, updated)
	return updated
}

func New(defaultValues Defaults, project project.Data) Defaults {
	defaultValues.Project = project
	defaultValues.triggersDefaulter = NewTriggersDefaulter()
	defaultValues.tasksDefaulter = NewTaskDefaulter()

	return defaultValues
}
