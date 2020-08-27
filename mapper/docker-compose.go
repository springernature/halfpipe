package mapper

import (
	"strings"

	"github.com/simonjohansson/yaml"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/config"
	"github.com/springernature/halfpipe/defaults"
	"github.com/springernature/halfpipe/manifest"
)

type dockerComposeMapper struct {
	fs afero.Afero
}

func NewDockerComposeMapper(fs afero.Afero) Mapper {
	return dockerComposeMapper{fs}
}

func (m dockerComposeMapper) Apply(original manifest.Manifest) (updated manifest.Manifest, err error) {
	if !original.FeatureToggles.DockerDecompose() {
		return original, nil
	}

	updated = original
	updated.Tasks = m.updateTasks(updated.Tasks)
	return updated, nil
}

func (m dockerComposeMapper) updateTasks(tasks manifest.TaskList) (updated manifest.TaskList) {
	for _, task := range tasks {
		switch task := task.(type) {
		case manifest.Parallel:
			task.Tasks = m.updateTasks(task.Tasks)
			updated = append(updated, task)
		case manifest.Sequence:
			task.Tasks = m.updateTasks(task.Tasks)
			updated = append(updated, task)
		case manifest.DeployCF:
			task.PrePromote = m.updateTasks(task.PrePromote)
			updated = append(updated, task)
		case manifest.DockerCompose:
			updated = append(updated, m.convertToRunTask(task))
		default:
			updated = append(updated, task)
		}
	}
	return updated
}

func (m dockerComposeMapper) convertToRunTask(dcTask manifest.DockerCompose) manifest.Task {
	dc, err := m.dockerComposeFile(dcTask.ComposeFile)
	if err != nil {
		return dcTask
	}

	if len(dc.Services) != 1 {
		return dcTask
	}

	service, ok := dc.Services[dcTask.Service]
	if !ok {
		return dcTask
	}

	//don't convert if working_dir is the parent dir
	for _, vol := range service.Volumes {
		v := strings.Split(vol, ":")
		if len(v) > 1 {
			if v[0] == ".." && v[1] == service.WorkingDir {
				return dcTask
			}
		}
	}

	//prefer the command set in halfpipe over docker-compose
	runScript := dcTask.Command
	if runScript == "" {
		runScript = service.Command
	}

	runTask := manifest.Run{
		Script: runScript,
		Docker: manifest.Docker{
			Image: service.Image,
		},
		Name:                   dcTask.Name,
		ManualTrigger:          dcTask.ManualTrigger,
		Vars:                   dcTask.Vars,
		SaveArtifacts:          dcTask.SaveArtifacts,
		RestoreArtifacts:       dcTask.RestoreArtifacts,
		SaveArtifactsOnFailure: dcTask.SaveArtifactsOnFailure,
		Retries:                dcTask.Retries,
		NotifyOnSuccess:        dcTask.NotifyOnSuccess,
		Notifications:          dcTask.Notifications,
		Timeout:                dcTask.Timeout,
		BuildHistory:           dcTask.BuildHistory,
	}

	//hmm, copied from defaulter :)
	if strings.HasPrefix(runTask.Docker.Image, config.DockerRegistry) {
		runTask.Docker.Username = defaults.DefaultValues.DockerUsername
		runTask.Docker.Password = defaults.DefaultValues.DockerPassword
	}

	return runTask
}

func (m dockerComposeMapper) dockerComposeFile(path string) (dc DockerCompose, err error) {
	content, err := m.fs.ReadFile(path)
	if err != nil {
		return dc, err
	}
	err = yaml.Unmarshal(content, &dc)
	return dc, err
}

type Service struct {
	Image      string
	Command    string
	Volumes    []string
	WorkingDir string `yaml:"working_dir"`
}

type DockerCompose struct {
	Services map[string]Service
}
