package manifest

import (
	"code.cloudfoundry.org/cli/util/manifest"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

const tagName = "secretAllowed"
var reservedKeyNames =  []string{"value"}

var UnsupportedSecretError = func(fieldName string) error {
	return fmt.Errorf("'%s' is not allowed to contain a secret", fieldName)
}

var InvalidSecretError = func(secret, fieldName string) error {
	return fmt.Errorf("'%s' at '%s' is not a valid key, must be in format of ((mapName.keyName)) with allowed characters [a-zA-Z0-9-_]", secret, fieldName)
}

var ReservedSecretNameError = func(secret, fieldName, reservedName string) error {
	return fmt.Errorf("'%s' at '%s' uses a reserved name ('%s') as key name. Reserved keywords are: %v", secret, fieldName, reservedName, reservedKeyNames)
}

type SecretValidator interface {
	Validate(Manifest) []error
}

type secretValidator struct {
}

func NewSecretValidator() SecretValidator {
	return secretValidator{}
}

func (s secretValidator) getRealFieldName(fieldName string, jsonTag string) string {
	if fieldName == "API" {
		return "api"
	}

	if jsonTag == "" || jsonTag == "-" || jsonTag == "omitempty" {
		return strings.ToLower(string(fieldName[0])) + fieldName[1:]
	}

	return strings.Split(jsonTag, ",")[0]
}

func (s secretValidator) validate(i interface{}, fieldName string, secretTag string, errs *[]error) {
	v := reflect.ValueOf(i)

	switch v.Type() {

	case reflect.TypeOf(Manifest{}),
		reflect.TypeOf(Repo{}),
		reflect.TypeOf(Run{}),
		reflect.TypeOf(Docker{}),
		reflect.TypeOf(DockerPush{}),
		reflect.TypeOf(DockerCompose{}),
		reflect.TypeOf(DeployCF{}),
		reflect.TypeOf(ConsumerIntegrationTest{}),
		reflect.TypeOf(DeployMLZip{}),
		reflect.TypeOf(DeployMLModules{}),
		reflect.TypeOf(ArtifactConfig{}),
		reflect.TypeOf(GitTrigger{}),
		reflect.TypeOf(TimerTrigger{}),
		reflect.TypeOf(DockerTrigger{}),
		reflect.TypeOf(PipelineTrigger{}):

		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			name := v.Type().Field(i).Name
			jsonTag := v.Type().Field(i).Tag.Get("json")
			secretTag := v.Type().Field(i).Tag.Get(tagName)

			var realFieldName string
			if fieldName == "" {
				realFieldName = s.getRealFieldName(name, jsonTag)
			} else {
				realFieldName = fmt.Sprintf("%s.%s", fieldName, s.getRealFieldName(name, jsonTag))
			}

			s.validate(field.Interface(), realFieldName, secretTag, errs)
		}

	case reflect.TypeOf(TaskList{}):
		for i, elem := range v.Interface().(TaskList) {
			realFieldName := fmt.Sprintf("%s[%d]", fieldName, i)
			s.validate(elem, realFieldName, secretTag, errs)
		}
	case reflect.TypeOf(TriggerList{}):
		for i, elem := range v.Interface().(TriggerList) {
			realFieldName := fmt.Sprintf("%s[%d]", fieldName, i)
			s.validate(elem, realFieldName, secretTag, errs)
		}
	case reflect.TypeOf(Parallel{}):
		for i, elem := range v.Interface().(Parallel).Tasks {
			realFieldName := fmt.Sprintf("%s[%d]", fieldName, i)
			s.validate(elem, realFieldName, secretTag, errs)
		}
	case reflect.TypeOf(Sequence{}):
		for i, elem := range v.Interface().(Sequence).Tasks {
			realFieldName := fmt.Sprintf("%s[%d]", fieldName, i)
			s.validate(elem, realFieldName, secretTag, errs)
		}
	case reflect.TypeOf([]string{"stringArray"}):
		for i, elem := range v.Interface().([]string) {
			realFieldName := fmt.Sprintf("%s[%d]", fieldName, i)
			s.validate(elem, realFieldName, secretTag, errs)
		}

	case reflect.TypeOf(FeatureToggles{}):
		for i, elem := range v.Interface().(FeatureToggles) {
			realFieldName := fmt.Sprintf("%s[%d]", fieldName, i)
			s.validate(elem, realFieldName, secretTag, errs)
		}

	case reflect.TypeOf(Vars{}):
		for key, value := range v.Interface().(Vars) {
			realKeyName := fmt.Sprintf("key %s[%s]", fieldName, key)
			s.validate(key, realKeyName, "false", errs)
			realValueName := fmt.Sprintf("%s[%s]", fieldName, key)
			s.validate(value, realValueName, secretTag, errs)
		}

	case reflect.TypeOf("string"):
		secret := v.Interface().(string)
		r := regexp.MustCompile(`\(\(.*\)\)`)

		if r.MatchString(secret) {
			if secretTag != "true" {
				*errs = append(*errs, UnsupportedSecretError(fieldName))
				return
			}

			splitSecret := strings.Split(secret, ".")
			if len(splitSecret) != 2 || !regexp.MustCompile(`^\(\([a-zA-Z0-9\-_\.]+\)\)$`).MatchString(secret) {
				*errs = append(*errs, InvalidSecretError(secret, fieldName))
				return
			}

			if len(splitSecret) == 2 {
				keyName := strings.ReplaceAll(splitSecret[1], ")", "")
				if s.IsReservedKeyName(keyName) {
					*errs = append(*errs, ReservedSecretNameError(secret, fieldName, keyName))
					return
				}
			}
		}

	case reflect.TypeOf(true), reflect.TypeOf(0), reflect.TypeOf(manifest.Application{}):
		// Stuff that we don't care about as they cannot contain secrets.
		return
	case reflect.TypeOf(Update{}):
		return
	case reflect.TypeOf(Platform("")):
		return
	case reflect.TypeOf(Notifications{}):
		notifications := v.Interface().(Notifications)

		for ni, success := range notifications.OnSuccess {
			fieldName := fmt.Sprintf("%s.on_success[%d]", fieldName, ni)
			s.validate(success, fieldName, secretTag, errs)
		}

		for ni, success := range notifications.OnFailure {
			fieldName := fmt.Sprintf("%s.on_failure[%d]", fieldName, ni)
			s.validate(success, fieldName, secretTag, errs)
		}
		return

	default:
		panic(fmt.Sprintf("Not implemented for %s", v.Type()))
	}

}

func (s secretValidator) Validate(man Manifest) (errors []error) {
	var errs []error
	s.validate(man, "", "", &errs)
	return errs
}

func (s secretValidator) IsReservedKeyName(keyName string) bool {
	for _, name := range reservedKeyNames {
		if keyName == name {
			return true
		}
	}
	return false
}