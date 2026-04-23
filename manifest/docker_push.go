package manifest

// docker-push builds a Docker image and pushes it to a docker registry. The
// image will be tagged with the latest tag, the gitref and pipeline version
// by default.
type DockerPush struct {
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
	// Optional display name.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// Docker image to build and push. Recommended format: eu.gcr.io/halfpipe-io/<team>/<image-name>.
	Image string `json:"image,omitempty" yaml:"image,omitempty" jsonschema:"required"`
	// Username for the target Docker registry.
	Username string `json:"username,omitempty" yaml:"username,omitempty" secretAllowed:"true"`
	// Password for the target Docker registry.
	Password string `json:"password,omitempty" yaml:"password,omitempty" secretAllowed:"true"`
	// Do not fail the build if critical vulnerabilities are found during image scanning.
	IgnoreVulnerabilities bool `json:"ignore_vulnerabilities,omitempty" yaml:"ignore_vulnerabilities,omitempty"`
	// Number of minutes a Trivy vulnerability scan is allowed to run before timing out.
	ScanTimeout int `json:"scan_timeout,omitempty" yaml:"scan_timeout,omitempty"`
	// Docker build-time variables (ARGs). Do not use for secrets - values are visible in docker history.
	Vars Vars `json:"vars,omitempty" yaml:"vars,omitempty" secretAllowed:"true"`
	// Docker build-time secrets, mounted securely during build.
	Secrets Vars `json:"secrets,omitempty" yaml:"secrets,omitempty" secretAllowed:"true"`
	// Restore artifacts saved by previous tasks.
	RestoreArtifacts bool `json:"restore_artifacts" yaml:"restore_artifacts,omitempty"`
	// Path to the Dockerfile, relative to the manifest. Defaults to Dockerfile.
	DockerfilePath string `json:"dockerfile_path,omitempty" yaml:"dockerfile_path,omitempty"`
	// Path to the folder to use as the Docker build context, relative to the manifest.
	BuildPath string `json:"build_path,omitempty" yaml:"build_path,omitempty"`
	// Deprecated: no longer used - safe to delete.
	Tag string `json:"tag,omitempty" yaml:"tag,omitempty" jsonschema_extras:"deprecated=true,deprecationMessage=no longer used - safe to delete"`
	// Target platforms to build for, e.g. linux/amd64, linux/arm64. Defaults to linux/amd64.
	Platforms []string `json:"platforms,omitempty" yaml:"platforms,omitempty"`
	// Enable layer caching to speed up builds by reusing layers from previous builds. Defaults to false.
	UseCache bool `json:"use_cache,omitempty" yaml:"use_cache,omitempty"`
	TaskBase `yaml:",inline"`
}

func (r DockerPush) SetNotifications(notifications Notifications) Task {
	r.Notifications = notifications
	return r
}

func (r DockerPush) SetTimeout(timeout string) Task {
	r.Timeout = timeout
	return r
}

func (r DockerPush) SetName(name string) Task {
	r.Name = name
	return r
}

func (r DockerPush) MarshalYAML() (any, error) {
	r.Type = "docker-push"
	return r, nil
}

func (r DockerPush) GetName() string {
	if r.Name == "" {
		return "docker-push"
	}
	return r.Name
}

func (r DockerPush) SetNotifyOnSuccess(notifyOnSuccess bool) Task {
	r.NotifyOnSuccess = notifyOnSuccess
	return r
}

func (r DockerPush) SavesArtifactsOnFailure() bool {
	return false
}

func (r DockerPush) SavesArtifacts() bool {
	return false
}

func (r DockerPush) ReadsFromArtifacts() bool {
	return r.RestoreArtifacts
}
