package manifest

import (
	"fmt"
	"github.com/simonjohansson/yaml"
	"github.com/springernature/halfpipe/linters/linterrors"
	"reflect"
	"regexp"
	"strings"
)

func Parse(manifestYaml string) (man Manifest, errors []error) {
	if err := yaml.Unmarshal([]byte(manifestYaml), &man); err != nil {
		errors = append(errors, err)
	}

	return
}

func getAllowedFieldsInType(t interface{}) (allowedFields []string) {
	reflected := reflect.ValueOf(t)
	for i := 0; i < reflected.NumField(); i++ {
		tag := reflected.Type().Field(i).Tag
		yamlTag := tag.Get("yaml")
		if yamlTag != "" && yamlTag != "-" {
			fieldName := strings.Split(yamlTag, ",")[0]
			allowedFields = append(allowedFields, fieldName)
		}
	}
	return allowedFields
}

func formatError(err error, t interface{}, prefix string) error {
	// To check if field does not exist in type
	fieldNotFoundRegex := regexp.MustCompile(`field (.*) not found in`)
	if fieldNotFoundRegex.MatchString(err.Error()) {
		fieldName := fieldNotFoundRegex.FindStringSubmatch(err.Error())[1]
		allowedFields := getAllowedFieldsInType(t)
		reasonText := fmt.Sprintf("must be one of '%s'", strings.Join(allowedFields, ", "))
		return linterrors.NewInvalidField(fmt.Sprintf("%s.%s", prefix, fieldName), reasonText)
	}

	// To check if we do something naughty with types
	if strings.Contains(err.Error(), "cannot unmarshal") {
		badTypeRegex := regexp.MustCompile(`line [0-9]+: (.*)`)
		typeErrorStr := badTypeRegex.FindStringSubmatch(err.Error())[1]
		return linterrors.NewInvalidField(prefix, typeErrorStr)
	}

	return fmt.Errorf("%s: %s", prefix, err)
}

func (t *TriggerList) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var triggers []map[string]interface{}
	if unmarshalErr := unmarshal(&triggers); unmarshalErr != nil {
		return unmarshalErr
	}

	for i, trigger := range triggers {
		prefix := fmt.Sprintf("triggers[%d]", i)
		yamlAgain, marshalErr := yaml.Marshal(trigger)
		if marshalErr != nil {
			return formatError(marshalErr, nil, prefix)
		}

		var typedTrigger Trigger
		var err error

		switch trigger["type"] {
		case "git":
			g := GitTrigger{}
			err = yaml.UnmarshalStrict(yamlAgain, &g)
			g.Type = ""
			typedTrigger = g
		case "docker":
			d := DockerTrigger{}
			err = yaml.UnmarshalStrict(yamlAgain, &d)
			d.Type = ""
			typedTrigger = d
		case "pipeline":
			p := PipelineTrigger{}
			err = yaml.UnmarshalStrict(yamlAgain, &p)
			p.Type = ""
			typedTrigger = p
		case "timer":
			t := TimerTrigger{}
			err = yaml.UnmarshalStrict(yamlAgain, &t)
			t.Type = ""
			typedTrigger = t
		default:
			triggerType := trigger["type"]
			if triggerType == nil {
				triggerType = ""
			}
			return linterrors.NewInvalidField(fmt.Sprintf("%s.type", prefix), fmt.Sprintf("was '%s' but must be one of 'git', 'pipeline', 'docker', 'cron'", triggerType))
		}

		if err != nil {
			return formatError(err, typedTrigger, prefix)
		}

		*t = append(*t, typedTrigger)
	}

	return nil
}

func (t *TaskList) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var tasks []map[string]interface{}
	if unmarshalErr := unmarshal(&tasks); unmarshalErr != nil {
		return unmarshalErr
	}

	for i, task := range tasks {
		prefix := fmt.Sprintf("tasks[%d]", i)
		yamlAgain, marshalErr := yaml.Marshal(task)
		if marshalErr != nil {
			return formatError(marshalErr, nil, prefix)
		}

		var typedTask Task
		var err error

		switch task["type"] {
		case "run":
			r := Run{}
			err = yaml.UnmarshalStrict(yamlAgain, &r)
			r.Type = ""
			typedTask = r
		case "docker-compose":
			d := DockerCompose{}
			err = yaml.UnmarshalStrict(yamlAgain, &d)
			d.Type = ""
			typedTask = d
		case "deploy-cf":
			c := DeployCF{}
			err = yaml.UnmarshalStrict(yamlAgain, &c)
			c.Type = ""
			typedTask = c
		case "docker-push":
			d := DockerPush{}
			err = yaml.UnmarshalStrict(yamlAgain, &d)
			d.Type = ""
			typedTask = d
		case "consumer-integration-test":
			c := ConsumerIntegrationTest{}
			err = yaml.UnmarshalStrict(yamlAgain, &c)
			c.Type = ""
			typedTask = c
		case "deploy-ml-zip":
			d := DeployMLZip{}
			err = yaml.UnmarshalStrict(yamlAgain, &d)
			d.Type = ""
			typedTask = d
		case "deploy-ml-modules":
			d := DeployMLModules{}
			err = yaml.UnmarshalStrict(yamlAgain, &d)
			d.Type = ""
			typedTask = d
		case "parallel":
			p := Parallel{}
			err = yaml.UnmarshalStrict(yamlAgain, &p)
			p.Type = ""
			typedTask = p
		case "sequence":
			p := Sequence{}
			err = yaml.UnmarshalStrict(yamlAgain, &p)
			p.Type = ""
			typedTask = p
		default:
			triggerType := task["type"]
			if triggerType == nil {
				triggerType = ""
			}
			return linterrors.NewInvalidField(fmt.Sprintf("%s.type", prefix), fmt.Sprintf("was '%s' but must be one of 'run', 'docker-compose', 'deploy-cf', 'docker-push', 'consumer-integration-test', 'deploy-ml-zip', 'deploy-ml-modules', 'parallel', 'sequence'", triggerType))
		}

		if err != nil {
			return formatError(err, typedTask, prefix)
		}

		*t = append(*t, typedTask)
	}

	return nil
}
