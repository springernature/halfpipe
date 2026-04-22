package manifest

// buildpack generates a container image using Cloud Native Buildpacks and
// publishes it to the Halfpipe registry. The task uses [Paketo Buildpacks]
// which is an implementation of the Cloud Native Buildpacks specification.
//
// [Paketo Buildpacks]: https://paketo.io
type Buildpack struct {
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
	// Optional display name.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// Docker image name to build and push. Format: eu.gcr.io/halfpipe-io/<team>/<image-name>.
	Image string `json:"image,omitempty" yaml:"image,omitempty" jsonschema:"required"`
	// Buildpack identifiers to use for building the image e.g. paketo-buildpacks/java.
	Buildpacks []string `json:"buildpacks" yaml:"buildpacks" jsonschema:"required"`
	// Paketo builder to use. Defaults to paketobuildpacks/builder-jammy-buildpackless-base.
	Builder string `json:"builder" yaml:"builder"`
	// Path to the application source code to build. Defaults to current directory.
	Path string `json:"path" yaml:"path"`
	// Restore artifacts saved by previous tasks.
	RestoreArtifacts bool `json:"restore_artifacts,omitempty" yaml:"restore_artifacts,omitempty"`
	// Environment variables passed to the pack build command.
	Vars     Vars `json:"vars,omitempty" yaml:"vars,omitempty" secretAllowed:"true"`
	TaskBase `yaml:",inline"`
}

func (p Buildpack) SetNotifications(notifications Notifications) Task {
	p.Notifications = notifications
	return p
}

func (p Buildpack) SetTimeout(timeout string) Task {
	p.Timeout = timeout
	return p
}

func (p Buildpack) SetName(name string) Task {
	p.Name = name
	return p
}

func (p Buildpack) MarshalYAML() (any, error) {
	p.Type = "buildpack"
	return p, nil
}

func (p Buildpack) GetName() string {
	if p.Name == "" {
		return "buildpack"
	}
	return p.Name
}

func (p Buildpack) SetNotifyOnSuccess(notifyOnSuccess bool) Task {
	p.NotifyOnSuccess = notifyOnSuccess
	return p
}

func (p Buildpack) SavesArtifactsOnFailure() bool {
	return false
}

func (p Buildpack) SavesArtifacts() bool {
	return false
}

func (p Buildpack) ReadsFromArtifacts() bool {
	return p.RestoreArtifacts
}
