package manifest

import "slices"

type Task interface {
	ReadsFromArtifacts() bool
	SavesArtifacts() bool
	SavesArtifactsOnFailure() bool
	SetTimeout(timeout string) Task
	GetName() string
	GetBase() TaskBase
	SetName(name string) Task
	SetNotifications(notifications Notifications) Task
	SetNotifyOnSuccess(notifyOnSuccess bool) Task
	MarshalYAML() (any, error) // To make sure type is always set when marshalling to yaml
}

type TaskList []Task

func (tl TaskList) SavesArtifacts() bool {
	return slices.ContainsFunc(tl, func(t Task) bool { return t.SavesArtifacts() })

}

func (tl TaskList) SavesArtifactsOnFailure() bool {
	return slices.ContainsFunc(tl, func(t Task) bool { return t.SavesArtifactsOnFailure() })
}

func (tl TaskList) UsesSlackNotifications() bool {
	for _, task := range tl {
		switch task := task.(type) {
		case Parallel:
			if task.Tasks.UsesSlackNotifications() {
				return true
			}
		case Sequence:
			if task.Tasks.UsesSlackNotifications() {
				return true
			}
		default:
			if len(task.GetBase().Notifications.Success.Slack()) > 0 || len(task.GetBase().Notifications.Failure.Slack()) > 0 {
				return true
			}
		}
	}
	return false
}

func (tl TaskList) UsesDockerPushWithCache() bool {
	return slices.ContainsFunc(tl.Flatten(), func(t Task) bool {
		if d, ok := t.(DockerPush); ok {
			return d.UseCache
		}
		return false
	})
}

func (tl TaskList) UsesTeamsNotifications() bool {
	for _, task := range tl {
		switch task := task.(type) {
		case Parallel:
			if task.Tasks.UsesTeamsNotifications() {
				return true
			}
		case Sequence:
			if task.Tasks.UsesTeamsNotifications() {
				return true
			}
		default:
			if len(task.GetBase().Notifications.Failure.Teams()) > 0 || len(task.GetBase().Notifications.Success.Teams()) > 0 {
				return true
			}
		}
	}
	return false
}

func (tl TaskList) Flatten() (updated TaskList) {
	for _, t := range tl {
		switch task := t.(type) {
		case DeployCF:
			copied := task
			copied.PrePromote = nil
			updated = append(updated, copied)
			updated = append(updated, task.PrePromote.Flatten()...)
		case Sequence:
			updated = append(updated, task.Tasks.Flatten()...)
		case Parallel:
			updated = append(updated, task.Tasks.Flatten()...)
		default:
			updated = append(updated, task)
		}
	}
	return
}

func (tl TaskList) GetTask(name string) Task {
	for _, t := range tl.Flatten() {
		if t.GetName() == name {
			return t
		}
	}
	return nil
}

func (tl TaskList) PreviousTaskNames(currentIndex int) []string {
	if currentIndex == 0 {
		return []string{}
	}
	return TaskNamesFromTask(tl[currentIndex-1])
}

func TaskNamesFromTask(t Task) (taskNames []string) {
	switch task := t.(type) {
	case Parallel:
		for _, subTask := range task.Tasks {
			taskNames = append(taskNames, TaskNamesFromTask(subTask)...)
		}
	case Sequence:
		lastTask := task.Tasks[len(task.Tasks)-1]
		taskNames = append(taskNames, TaskNamesFromTask(lastTask)...)
	default:
		taskNames = append(taskNames, task.GetName())
	}

	return taskNames
}

type TaskBase struct {
	// Task must be triggered manually (Concourse only).
	ManualTrigger bool `json:"manual_trigger,omitempty" yaml:"manual_trigger,omitempty" jsonschema:"default=false"`
	// Number of times to retry the task if it fails.
	Retries int `json:"retries,omitempty" yaml:"retries,omitempty" jsonschema:"default=0"`
	// Deprecated: use notifications instead.
	NotifyOnSuccess bool `json:"notify_on_success,omitempty" yaml:"notify_on_success,omitempty" jsonschema:"default=false" jsonschema_extras:"deprecated=true,deprecationMessage=use notifications instead"`
	// Notification channels for this task.
	Notifications Notifications `json:"notifications" yaml:"notifications,omitempty"`
	// Timeout duration for the task. If exceeded the task fails.
	Timeout string `json:"timeout,omitempty" yaml:"timeout,omitempty" jsonschema:"default=1h"`
	// Number of build logs to retain (Concourse only).
	BuildHistory int `json:"build_history,omitempty" yaml:"build_history,omitempty" jsonschema:"default=20"`
}

func (t TaskBase) GetBase() TaskBase {
	return t
}
