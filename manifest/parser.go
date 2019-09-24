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
		task, err := unmarshalTask(i, rawTasks[i], typedObject.Type)
		if err != nil {
			return err
		}
		*t = append(*t, task)
	}

	return nil
}

func (t *TriggerList) UnmarshalJSON(b []byte) error {
	// first get a raw array
	var rawTrigger []json.RawMessage
	if err := json.Unmarshal(b, &rawTrigger); err != nil {
		return err
	}

	// then just read the type field
	var objectsWithType []objectWithType
	if err := json.Unmarshal(b, &objectsWithType); err != nil {
		return err
	}

	// should have 2 arrays the same length..
	if len(rawTrigger) != len(objectsWithType) {
		return fmt.Errorf("error parsing trigger")
	}

	// loop through and use the Type field to unmarshal into the correct type of Task
	for i, typedObject := range objectsWithType {
		trigger, err := unmarshalTrigger(i, rawTrigger[i], typedObject.Type)
		if err != nil {
			return err
		}
		*t = append(*t, trigger)
	}

	return nil
}

func (t *ParallelGroup) UnmarshalJSON(b []byte) error {
	var rawTask json.RawMessage
	if err := json.Unmarshal(b, &rawTask); err != nil {
		return err
	}

	*t = ParallelGroup(rawTask)
	return nil
}

func unmarshalTask(taskIndex int, rawTask json.RawMessage, taskType string) (task Task, err error) {

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
			return nil, err
		}
		t.Type = ""
		task = t
	case "deploy-cf":
		t := DeployCF{}
		if err := unmarshal(rawTask, &t, taskIndex); err != nil {
			return nil, err
		}
		t.Type = ""
		task = t
	case "docker-push":
		t := DockerPush{}
		if err := unmarshal(rawTask, &t, taskIndex); err != nil {
			return nil, err
		}
		t.Type = ""
		task = t
	case "docker-compose":
		t := DockerCompose{}
		if err := unmarshal(rawTask, &t, taskIndex); err != nil {
			return nil, err
		}
		t.Type = ""
		task = t
	case "consumer-integration-test":
		t := ConsumerIntegrationTest{}
		if err := unmarshal(rawTask, &t, taskIndex); err != nil {
			return nil, err
		}
		t.Type = ""
		task = t
	case "deploy-ml-zip":
		t := DeployMLZip{}
		if err := unmarshal(rawTask, &t, taskIndex); err != nil {
			return nil, err
		}
		t.Type = ""
		task = t
	case "deploy-ml-modules":
		t := DeployMLModules{}
		if err := unmarshal(rawTask, &t, taskIndex); err != nil {
			return nil, err
		}
		t.Type = ""
		task = t
	case "parallel":
		t := Parallel{}
		if err := unmarshal(rawTask, &t, taskIndex); err != nil {
			return nil, err
		}
		t.Type = ""
		task = t
	case "seq":
		t := Seq{}
		if err := unmarshal(rawTask, &t, taskIndex); err != nil {
			return nil, err
		}
		t.Type = ""
		task = t

	default:
		err = errors.NewInvalidField("task", fmt.Sprintf("tasks.task[%v] unknown type '%s'. Must be one of 'run', 'docker-compose', 'deploy-cf', 'docker-push', 'consumer-integration-test', 'parallel', 'seq'", taskIndex, taskType))
	}

	return
}

func unmarshalTrigger(triggerIndex int, rawTrigger json.RawMessage, triggerType string) (trigger Trigger, err error) {

	unmarshal := func(rawTrigger json.RawMessage, t Trigger, index int) error {
		decoder := json.NewDecoder(bytes.NewReader(rawTrigger))
		decoder.DisallowUnknownFields()
		if jsonErr := decoder.Decode(t); jsonErr != nil {
			return errors.NewInvalidField("trigger", fmt.Sprintf("triggers.trigger[%v] %s", index, jsonErr.Error()))
		}
		return nil
	}

	// unmarshal into the correct type of Task
	switch triggerType {
	case "git":
		t := GitTrigger{}
		if err := unmarshal(rawTrigger, &t, triggerIndex); err != nil {
			return nil, err
		}
		t.Type = ""
		trigger = t
	case "timer":
		t := TimerTrigger{}
		if err := unmarshal(rawTrigger, &t, triggerIndex); err != nil {
			return nil, err
		}
		t.Type = ""
		trigger = t
	case "docker":
		t := DockerTrigger{}
		if err := unmarshal(rawTrigger, &t, triggerIndex); err != nil {
			return nil, err
		}
		t.Type = ""
		trigger = t
	default:
		err = errors.NewInvalidField("task", fmt.Sprintf("triggers.trigger[%v] unknown type '%s'. Must be one of 'git', 'cron'", triggerIndex, triggerType))
	}

	return
}
