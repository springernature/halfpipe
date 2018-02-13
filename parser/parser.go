package parser

import (
	"encoding/json"

	"fmt"

	"github.com/ghodss/yaml"
	"github.com/springernature/halfpipe/errors"
	. "github.com/springernature/halfpipe/model"
)

func Parse(manifestYaml string) (man Manifest, errs []error) {
	addError := func(e error) {
		errs = append(errs, e)
	}

	if err := yaml.Unmarshal([]byte(manifestYaml), &man); err != nil {
		addError(errors.NewParseError(err.Error()))
		return
	}

	var rawTasks struct {
		Tasks []json.RawMessage
	}
	if err := yaml.Unmarshal([]byte(manifestYaml), &rawTasks); err != nil {
		addError(errors.NewParseError(err.Error()))
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

		taskName := struct {
			Name string
		}{}

		if err := json.Unmarshal(rawTask, &taskName); err != nil {
			addError(errors.NewInvalidField("task", fmt.Sprintf("task %v %s", i+1, err.Error())))
			return
		}

		switch taskName.Name {
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
			addError(errors.NewInvalidField("task", fmt.Sprintf("task %v is missing name field", i+1)))
		default:
			addError(errors.NewInvalidField("task", fmt.Sprintf("task %v has unknown name '%s'", i+1, taskName.Name)))
		}
	}
	return
}
