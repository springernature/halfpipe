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

type OutputDefaulter interface {
	Apply(original manifest.Manifest) (updated manifest.Manifest)
}

type BuildHistoryDefaulter interface {
	Apply(original manifest.TaskList, defaults Defaults) (updated manifest.TaskList)
}

type CFSnPaaS struct {
	Username string
	Password string
	Org      string
	API      string
}

type CFDefaults struct {
	ManifestPath string
	SnPaaS       CFSnPaaS
	TestDomains  map[string]string
	Version      string
}

type KateeDefaults struct {
	VelaManifest string
	Tag          string
}

type DockerDefaults struct {
	Username          string
	Password          string
	FilePath          string
	ComposeFile       string
	ComposeService    string
	ImageScanSeverity string
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

	ShallowClone   bool
	RepoPrivateKey string
	CF             CFDefaults
	Katee          KateeDefaults
	Docker         DockerDefaults
	Artifactory    ArtifactoryDefaults
	Concourse      ConcourseDefaults
	MarkLogic      MarkLogicDefaults

	Timeout string

	triggersDefaulter TriggersDefaulter
	tasksDefaulter    TasksDefaulter
	outputDefaulter   OutputDefaulter
}

func (d Defaults) Apply(original manifest.Manifest) (updated manifest.Manifest) {
	updated = d.outputDefaulter.Apply(original)
	updated.Triggers = d.triggersDefaulter.Apply(updated.Triggers, d, original)
	updated.Tasks = d.tasksDefaulter.Apply(updated.Tasks, d, updated)
	return updated
}

func New(defaultValues Defaults, project project.Data) Defaults {
	defaultValues.Project = project
	defaultValues.triggersDefaulter = NewTriggersDefaulter()
	defaultValues.tasksDefaulter = NewTaskDefaulter()
	defaultValues.outputDefaulter = NewOutputDefaulter()

	return defaultValues
}
