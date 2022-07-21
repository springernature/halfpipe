package manifest

type DeployKatee struct {
	Type              string
	Name              string `yaml:"name,omitempty"`
	ApplicationName   string `yaml:"applicationName,omitempty"`
	BuildVersion      string `yaml:"buildVersion,omitempty"`
	BuildArgs         string `yaml:"buildArgs,omitempty"`
	Dockerfile        string `yaml:"dockerfile,omitempty"`
	ImageName         string `yaml:"imageName,omitempty"`
	SlackChannel      string `yaml:"slackChannel,omitempty"`
	ImageScanSeverity string `yaml:"imageScanSeverity,omitempty"`
	ApplicationRoot   string `yaml:"applicationRoot,omitempty"`
	Environment       string `yaml:"environment,omitempty"`
	Url               string `yaml:"url,omitempty"`
	VaultEnvVars      string `yaml:"vaultEnvVars,omitempty"`
	Timeout           string
}

func (d DeployKatee) ReadsFromArtifacts() bool {
	return false
}

func (d DeployKatee) GetAttempts() int {
	return 2
}

func (d DeployKatee) SavesArtifacts() bool {
	return false
}

func (d DeployKatee) SavesArtifactsOnFailure() bool {
	return false
}

func (d DeployKatee) IsManualTrigger() bool {
	return false
}

func (d DeployKatee) NotifiesOnSuccess() bool {
	return false
}

func (d DeployKatee) GetTimeout() string {
	d.Timeout = ""
	return d.Timeout
}

func (d DeployKatee) SetTimeout(timeout string) Task {
	d.Timeout = timeout
	return d
}

func (r DeployKatee) GetName() string {
	if r.Name == "" {
		return "deploy-katee"
	}
	return r.Name
}

func (r DeployKatee) SetName(name string) Task {
	r.Name = name
	return r
}

func (d DeployKatee) GetNotifications() Notifications {
	return Notifications{
		OnFailure: []string{d.SlackChannel},
	}
}

func (d DeployKatee) SetNotifications(notifications Notifications) Task {
	d.SlackChannel = notifications.OnFailure[0]
	return d
}

func (d DeployKatee) GetBuildHistory() int {
	return 1
}

func (d DeployKatee) SetBuildHistory(buildHistory int) Task {
	return d
}

func (d DeployKatee) GetSecrets() map[string]string {
	return map[string]string{}
}

func (d DeployKatee) MarshalYAML() (interface{}, error) {
	d.Type = "deploy-katee"
	return d, nil
}

//func (r DeployKatee) GetSecrets() map[string]string {
//	return findSecrets(map[string]string{})
//}
//
//func (r DeployKatee) GetNotifications() Notifications {
//	return r.Notifications
//}
//
//func (r DeployKatee) SetNotifications(notifications Notifications) Task {
//	r.Notifications = notifications
//	return r
//}
//
//func (r DeployKatee) SetTimeout(timeout string) Task {
//	r.Timeout = timeout
//	return r
//}
//
//func (r DeployKatee) SetName(name string) Task {
//	r.Name = name
//	return r
//}
//
//func (r DeployKatee) MarshalYAML() (interface{}, error) {
//	r.Type = "deploy-cf"
//	return r, nil
//}
//

//
//func (r DeployKatee) GetTimeout() string {
//	return r.Timeout
//}
//
//func (r DeployKatee) NotifiesOnSuccess() bool {
//	return r.NotifyOnSuccess
//}
//
//func (r DeployKatee) SavesArtifactsOnFailure() bool {
//	for _, task := range r.PrePromote {
//		if task.SavesArtifactsOnFailure() {
//			return true
//		}
//	}
//	return false
//}
//
//func (r DeployKatee) IsManualTrigger() bool {
//	return r.ManualTrigger
//}
//
//func (r DeployKatee) SavesArtifacts() bool {
//	return false
//}
//
//func (r DeployKatee) ReadsFromArtifacts() bool {
//	if r.DeployArtifact != "" || strings.HasPrefix(r.Manifest, "../artifacts/") {
//		return true
//	}
//
//	for _, pp := range r.PrePromote {
//		if pp.ReadsFromArtifacts() {
//			return true
//		}
//	}
//	return false
//}
//
//func (r DeployKatee) GetAttempts() int {
//	return 2 + r.Retries
//}
