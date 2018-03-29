package manifest

import (
	"bytes"
	"strings"

	"fmt"

	"encoding/json"

	"github.com/ghodss/yaml"
	"github.com/springernature/halfpipe/linters/errors"
)

func Parse(manifestYaml string) (man Manifest, errs []error) {
	addError := func(e error) {
		errs = append(errs, e)
	}

	if es := unmarshalAsJSON([]byte(manifestYaml), &man); len(es) > 0 {
		for _, err := range es {
			addError(NewParseError(err.Error()))
		}
		return
	}
	return
}

// convert YAML to JSON because JSON parser gives more control that we need to unmarshal into tasks
func unmarshalAsJSON(yml []byte, out *Manifest) []error {
	js, err := yaml.YAMLToJSON(yml)
	if err != nil {
		return []error{err}
	}

	decoder := json.NewDecoder(bytes.NewReader(js))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(out); err != nil {
		msg := strings.Replace(err.Error(), "json: ", "", -1)
		return []error{fmt.Errorf(msg)}
	}
	return nil
}

type objectWithType struct {
	Type string
}

func (t *TaskList) UnmarshalJSON(b []byte) error {
	// first get a raw array
	var rawTasks []json.RawMessage
	if err := json.Unmarshal(b, &rawTasks); err != nil {
		return err
	}

	// then just read the type field
	var objectsWithType []objectWithType
	if err := json.Unmarshal(b, &objectsWithType); err != nil {
		return err
	}

	// should have 2 arrays the same length..
	if len(rawTasks) != len(objectsWithType) {
		return fmt.Errorf("error parsing tasks")
	}

	// loop through and use the Type field to unmarshal into the correct type of Task
	for i, typedObject := range objectsWithType {
		if err := unmarshalTask(t, i, rawTasks[i], typedObject.Type); err != nil {
			return err
		}
	}
	return nil
}

func unmarshalTask(taskList *TaskList, taskIndex int, rawTask json.RawMessage, taskType string) error {

	unmarshal := func(rawTask json.RawMessage, t Task, index int) error {
		decoder := json.NewDecoder(bytes.NewReader(rawTask))
		decoder.DisallowUnknownFields()
		if jsonErr := decoder.Decode(t); jsonErr != nil {
			return errors.NewInvalidField("task", fmt.Sprintf("tasks.task[%v] %s", index, jsonErr.Error()))
		}
		return nil
	}

	// unmarshal into the correct type of Task
	switch taskType {
	case "run":
		t := Run{}
		if err := unmarshal(rawTask, &t, taskIndex); err != nil {
			return err
		}
		t.Type = ""
		*taskList = append(*taskList, t)
	case "deploy-cf":
		t := DeployCF{}
		if err := unmarshal(rawTask, &t, taskIndex); err != nil {
			return err
		}
		t.Type = ""
		*taskList = append(*taskList, t)
	case "docker-push":
		t := DockerPush{}
		if err := unmarshal(rawTask, &t, taskIndex); err != nil {
			return err
		}
		t.Type = ""
		*taskList = append(*taskList, t)
	case "docker-compose":
		t := DockerCompose{}
		if err := unmarshal(rawTask, &t, taskIndex); err != nil {
			return err
		}
		t.Type = ""
		*taskList = append(*taskList, t)

	default:
		return errors.NewInvalidField("task", fmt.Sprintf("tasks.task[%v] unknown type '%s'. Must be one of 'run', 'docker-compose', 'deploy-cf', 'docker-push'", taskIndex, taskType))
	}

	return nil
}
