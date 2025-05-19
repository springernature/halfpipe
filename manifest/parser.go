package manifest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sigs.k8s.io/yaml"
	"strings"
)

func Parse(manifestYaml string) (Manifest, []error) {
	var man Manifest
	errs := unmarshalAsJSON([]byte(manifestYaml), &man)
	return man, errs
}

// convert YAML to JSON because JSON parser gives more control that we need to unmarshal into tasks
func unmarshalAsJSON(yml []byte, out *Manifest) []error {
	js, err := yaml.YAMLToJSONStrict(yml)
	if err != nil {
		if strings.Contains(err.Error(), "already set") {
			msg := strings.Replace(err.Error(), "already set in map", "is duplicated.", -1)
			return []error{fmt.Errorf("%s", msg)}
		}
		return []error{err}
	}

	decoder := json.NewDecoder(bytes.NewReader(js))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(out); err != nil {
		msg := strings.Replace(err.Error(), "json: ", "", -1)
		return []error{fmt.Errorf("%s", msg)}
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

func (v *Vars) UnmarshalJSON(b []byte) error {
	var rawVars map[string]interface{}
	if err := json.Unmarshal(b, &rawVars); err != nil {
		return err
	}

	var tmpVars Vars

	if len(rawVars) > 0 {
		tmpVars = make(Vars)
		for key, val := range rawVars {
			tmpVars[key] = fmt.Sprintf("%v", val)
		}
	}

	*v = tmpVars
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

func unmarshalTask(taskIndex int, rawTask json.RawMessage, taskType string) (task Task, err error) {
	decoder := json.NewDecoder(bytes.NewReader(rawTask))
	decoder.DisallowUnknownFields()

	unmarshal := func(t Task) error {
		if jsonErr := decoder.Decode(t); jsonErr != nil {
			return fmt.Errorf("tasks[%v] : %w", taskIndex, jsonErr)
		}
		return nil
	}

	// unmarshal into the correct type of Task
	switch taskType {
	case "run":
		t := Run{}
		err = unmarshal(&t)
		t.Type = ""
		task = t
	case "deploy-cf":
		t := DeployCF{}
		err = unmarshal(&t)
		t.Type = ""
		task = t
	case "deploy-katee":
		t := DeployKatee{}
		err = unmarshal(&t)
		t.Type = ""
		task = t
	case "docker-push":
		t := DockerPush{}
		err = unmarshal(&t)
		t.Type = ""
		task = t
	case "docker-compose":
		t := DockerCompose{}
		err = unmarshal(&t)
		t.Type = ""
		task = t
	case "consumer-integration-test":
		t := ConsumerIntegrationTest{}
		err = unmarshal(&t)
		//default use_covenant to true
		if !strings.Contains(string(rawTask), `"use_covenant"`) {
			t.UseCovenant = true
		}
		t.Type = ""
		task = t
	case "deploy-ml-zip":
		t := DeployMLZip{}
		err = unmarshal(&t)
		t.Type = ""
		task = t
	case "deploy-ml-modules":
		t := DeployMLModules{}
		err = unmarshal(&t)
		t.Type = ""
		task = t
	case "parallel":
		t := Parallel{}
		err = unmarshal(&t)
		t.Type = ""
		task = t
	case "pack":
		t := Pack{}
		err = unmarshal(&t)
		t.Type = ""
		task = t
	case "sequence":
		t := Sequence{}
		err = unmarshal(&t)
		t.Type = ""
		task = t

	default:
		err = fmt.Errorf("tasks[%v] unknown type '%s'. Must be one of 'run', 'docker-compose', 'deploy-cf', 'docker-push', 'consumer-integration-test', 'pack', 'parallel', 'sequence'", taskIndex, taskType)
	}

	return task, err
}

func unmarshalTrigger(triggerIndex int, rawTrigger json.RawMessage, triggerType string) (trigger Trigger, err error) {
	decoder := json.NewDecoder(bytes.NewReader(rawTrigger))
	decoder.DisallowUnknownFields()

	unmarshal := func(t Trigger) error {
		if jsonErr := decoder.Decode(t); jsonErr != nil {
			return fmt.Errorf("triggers.trigger[%v] : %w", triggerIndex, jsonErr)
		}
		return nil
	}

	// unmarshal into the correct type of Trigger
	switch triggerType {
	case "git":
		t := GitTrigger{ShallowDefined: strings.Contains(string(rawTrigger), `"shallow":`)}
		err = unmarshal(&t)
		t.Type = ""
		trigger = t
	case "timer":
		t := TimerTrigger{}
		err = unmarshal(&t)
		t.Type = ""
		trigger = t
	case "docker":
		t := DockerTrigger{}
		err = unmarshal(&t)
		t.Type = ""
		trigger = t
	case "pipeline":
		t := PipelineTrigger{}
		err = unmarshal(&t)
		t.Type = ""
		trigger = t
	default:
		err = fmt.Errorf("triggers[%v] unknown type '%s'. Must be one of 'git', 'cron', 'docker', 'pipeline'", triggerIndex, triggerType)
	}

	return trigger, err
}
