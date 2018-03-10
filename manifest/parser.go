package manifest

import (
	"github.com/ghodss/yaml"

	"fmt"

	"encoding/json"

	"github.com/springernature/halfpipe/linters/errors"
)

func Parse(manifestYaml string) (man Manifest, errs []error) {
	addError := func(e error) {
		errs = append(errs, e)
	}

	if err := yaml.Unmarshal([]byte(manifestYaml), &man); err != nil {
		addError(NewParseError(err.Error()))
		return
	}

	var rawTasks struct {
		Tasks []json.RawMessage
	}
	if err := yaml.Unmarshal([]byte(manifestYaml), &rawTasks); err != nil {
		addError(NewParseError(err.Error()))
		return
	}

	parseTask := func(rawTask json.RawMessage, t Task, index int) error {
		if err := json.Unmarshal(rawTask, t); err != nil {
			addError(errors.NewInvalidField("task", fmt.Sprintf("task %v %s", index+1, err.Error())))
			return err
		}
		return nil
	}

	for i, rawTask := range rawTasks.Tasks {
		// first unmarshall into struct with just 'Type' field
		taskType := struct {
			Type string
		}{}

		if err := json.Unmarshal(rawTask, &taskType); err != nil {
			addError(errors.NewInvalidField("task", fmt.Sprintf("task %v %s", i+1, err.Error())))
			return
		}

		// then use the value of 'Type' to unmarshall into the correct Task
		switch taskType.Type {
		case "run":
			t := Run{}
			if err := parseTask(rawTask, &t, i); err == nil {
				man.Tasks = append(man.Tasks, t)
			}
		case "deploy-cf":
			t := DeployCF{}
			if err := parseTask(rawTask, &t, i); err == nil {
				man.Tasks = append(man.Tasks, t)
			}
		case "docker-push":
			t := DockerPush{}
			if err := parseTask(rawTask, &t, i); err == nil {
				man.Tasks = append(man.Tasks, t)
			}
		case "":
			addError(errors.NewInvalidField("task", fmt.Sprintf("task %v is missing field 'type'", i+1)))
		default:
			addError(errors.NewInvalidField("task", fmt.Sprintf("task %v has unknown type '%s'. Must be one of 'run', 'deploy-cf', 'docker-push'", i+1, taskType.Type)))
		}
	}
	return
}
