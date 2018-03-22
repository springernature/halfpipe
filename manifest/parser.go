package manifest

import (
	"bytes"
	"strings"

	"github.com/ghodss/yaml"

	"fmt"

	"encoding/json"

	"github.com/springernature/halfpipe/linters/errors"
	"github.com/xeipuuv/gojsonschema"
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

	//delete the tasks, they are just map[string]interface{}, we can do better
	man.Tasks = nil

	var rawTasks struct {
		Tasks []json.RawMessage
	}
	if err := yaml.Unmarshal([]byte(manifestYaml), &rawTasks); err != nil {
		addError(NewParseError(err.Error()))
		return
	}

	for i, rawTask := range rawTasks.Tasks {
		if err := unmarshalTask(rawTask, i, &man.Tasks); err != nil {
			addError(NewParseError(err.Error()))
			return
		}
	}
	return
}

func unmarshalTask(rawTask json.RawMessage, taskIndex int, taskArray *[]Task) (err error) {

	parseTask := func(rawTask json.RawMessage, t Task, index int) error {
		if jsonErr := json.Unmarshal(rawTask, t); jsonErr != nil {
			return errors.NewInvalidField("task", fmt.Sprintf("task %v %s", index+1, jsonErr.Error()))
		}
		return nil
	}

	// first unmarshall into struct with just 'Type' field
	taskType := struct {
		Type string
	}{}

	if unmarshalErr := json.Unmarshal(rawTask, &taskType); unmarshalErr != nil {
		err = errors.NewInvalidField("task", fmt.Sprintf("task %v %s", taskIndex+1, unmarshalErr.Error()))
		return
	}

	// then use the value of 'Type' to unmarshall into the correct Task
	switch taskType.Type {
	case "run":
		t := Run{}
		if parseErr := parseTask(rawTask, &t, taskIndex); parseErr == nil {
			t.Type = "" // delete the type it's just for parsing
			*taskArray = append(*taskArray, t)
		}
	case "deploy-cf":
		t := DeployCF{}
		if parseErr := parseTask(rawTask, &t, taskIndex); parseErr == nil {
			t.Type = "" // delete the type it's just for parsing
			if len(t.PrePromote) > 0 {
				t.PrePromote = nil
				var rawTasks struct {
					PrePromote []json.RawMessage `json:"pre_promote"`
				}
				if unmarshalErr := yaml.Unmarshal([]byte(rawTask), &rawTasks); unmarshalErr != nil {
					return NewParseError(unmarshalErr.Error())
				}

				for i, rawTask := range rawTasks.PrePromote {
					if unmarshalErr := unmarshalTask(rawTask, i, &t.PrePromote); unmarshalErr != nil {
						return NewParseError(unmarshalErr.Error())
					}
				}
			}
			*taskArray = append(*taskArray, t)
		}
	case "docker-push":
		t := DockerPush{}
		if parseErr := parseTask(rawTask, &t, taskIndex); parseErr == nil {
			t.Type = "" // delete the type it's just for parsing
			*taskArray = append(*taskArray, t)
		}
	case "docker-compose":
		t := DockerCompose{}
		if parseErr := parseTask(rawTask, &t, taskIndex); parseErr == nil {
			t.Type = "" // delete the type it's just for parsing
			*taskArray = append(*taskArray, t)
		}
	case "":
		err = errors.NewInvalidField("task", fmt.Sprintf("task %v is missing field 'type'", taskIndex+1))
	default:
		err = errors.NewInvalidField("task", fmt.Sprintf("task %v has unknown type '%s'. Must be one of 'run', 'docker-compose', 'deploy-cf', 'docker-push'", taskIndex+1, taskType.Type))
	}
	return
}

// convert YAML to JSON because JSON parser gives more control that we need to unmarshal into tasks
// and bonus feature - we can validate against a JSON schema
func unmarshalAsJSON(yml []byte, out *Manifest) []error {
	js, err := yaml.YAMLToJSON(yml)
	if err != nil {
		return []error{err}
	}

	jsonReader := bytes.NewReader(js)
	decoder := json.NewDecoder(jsonReader)

	if errs := validate(js); len(errs) > 0 {
		return errs
	}
	if err := decoder.Decode(out); err != nil {
		msg := strings.Replace(err.Error(), "json: ", "", -1)
		return []error{fmt.Errorf("error parsing YAML: %v", msg)}
	}
	return nil
}

func validate(jsonManifest []byte) []error {
	schemaLoader := gojsonschema.NewStringLoader(jsonSchema)
	documentLoader := gojsonschema.NewBytesLoader(jsonManifest)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return []error{err}
	}
	if !result.Valid() {
		var errs []error
		for _, err := range result.Errors() {
			//tidy up the errors a bit
			msg := strings.Replace(err.String(), "Must validate at least one schema (anyOf)", "Task does not have required fields:", -1)
			msg = strings.Replace(msg, "Invalid type. Expected: object, given: null", "Not a valid YAML file", -1)
			msg = strings.Replace(msg, "must be one of the following: \"run\"", "must be one of the following: 'run', 'docker-compose', 'docker-push' or 'cf-deploy'", -1)
			ignore := strings.Contains(msg, ".type: Does not match pattern")
			if !ignore {
				errs = append(errs, fmt.Errorf(msg))
			}
		}
		return errs
	}
	return nil
}
