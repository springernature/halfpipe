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
				t.Type = "" // delete the type it's just for parsing
				man.Tasks = append(man.Tasks, t)
			}
		case "deploy-cf":
			t := DeployCF{}
			if err := parseTask(rawTask, &t, i); err == nil {
				t.Type = "" // delete the type it's just for parsing
				man.Tasks = append(man.Tasks, t)
			}
		case "docker-push":
			t := DockerPush{}
			if err := parseTask(rawTask, &t, i); err == nil {
				t.Type = "" // delete the type it's just for parsing
				man.Tasks = append(man.Tasks, t)
			}
		case "docker-compose":
			t := DockerCompose{}
			if err := parseTask(rawTask, &t, i); err == nil {
				t.Type = "" // delete the type it's just for parsing
				man.Tasks = append(man.Tasks, t)
			}
		case "":
			addError(errors.NewInvalidField("task", fmt.Sprintf("task %v is missing field 'type'", i+1)))
		default:
			addError(errors.NewInvalidField("task", fmt.Sprintf("task %v has unknown type '%s'. Must be one of 'run', 'docker-compose', 'deploy-cf', 'docker-push'", i+1, taskType.Type)))
		}
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
			msg := strings.Replace(err.String(), "Must validate at least one schema (anyOf)", "Invalid task", -1)
			ignore := strings.Contains(msg, ".type: Does not match pattern")
			if !ignore {
				errs = append(errs, fmt.Errorf(msg))
			}
		}
		return errs
	}
	return nil
}
