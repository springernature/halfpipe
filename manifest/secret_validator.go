package manifest

import (
	"code.cloudfoundry.org/cli/util/manifest"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

const tagName = "secretAllowed"

var reservedKeyNames = []string{"value"}

var UnsupportedSecretError = func(fieldName string) error {
	return fmt.Errorf("'%s' is not allowed to contain a secret", fieldName)
}

var InvalidSecretConcourseError = func(secret, fieldName string) error {
	return fmt.Errorf("'%s' at '%s' is not a valid key, must be in format of ((mapName.keyName))", secret, fieldName)
}

var InvalidSecretActionsError = func(secret, fieldName string) error {
	return fmt.Errorf("'%s' at '%s' is not a valid key, must be in format of ((mapName.keyName)) or ((/path/to/mapName keyName))", secret, fieldName)
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

func (s secretValidator) validate(i interface{}, fieldName string, secretTag string, errs *[]error, platform Platform) {
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

			s.validate(field.Interface(), realFieldName, secretTag, errs, platform)
		}

	case reflect.TypeOf(TaskList{}):
		for i, elem := range v.Interface().(TaskList) {
			realFieldName := fmt.Sprintf("%s[%d]", fieldName, i)
			s.validate(elem, realFieldName, secretTag, errs, platform)
		}
	case reflect.TypeOf(TriggerList{}):
		for i, elem := range v.Interface().(TriggerList) {
			realFieldName := fmt.Sprintf("%s[%d]", fieldName, i)
			s.validate(elem, realFieldName, secretTag, errs, platform)
		}
	case reflect.TypeOf(Parallel{}):
		for i, elem := range v.Interface().(Parallel).Tasks {
			realFieldName := fmt.Sprintf("%s[%d]", fieldName, i)
			s.validate(elem, realFieldName, secretTag, errs, platform)
		}
	case reflect.TypeOf(Sequence{}):
		for i, elem := range v.Interface().(Sequence).Tasks {
			realFieldName := fmt.Sprintf("%s[%d]", fieldName, i)
			s.validate(elem, realFieldName, secretTag, errs, platform)
		}
	case reflect.TypeOf([]string{"stringArray"}):
		for i, elem := range v.Interface().([]string) {
			realFieldName := fmt.Sprintf("%s[%d]", fieldName, i)
			s.validate(elem, realFieldName, secretTag, errs, platform)
		}

	case reflect.TypeOf(FeatureToggles{}):
		for i, elem := range v.Interface().(FeatureToggles) {
			realFieldName := fmt.Sprintf("%s[%d]", fieldName, i)
			s.validate(elem, realFieldName, secretTag, errs, platform)
		}

	case reflect.TypeOf(Vars{}):
		for key, value := range v.Interface().(Vars) {
			realKeyName := fmt.Sprintf("key %s[%s]", fieldName, key)
			s.validate(key, realKeyName, "false", errs, "")
			realValueName := fmt.Sprintf("%s[%s]", fieldName, key)
			s.validate(value, realValueName, secretTag, errs, platform)
		}

	case reflect.TypeOf("string"):
		secret := v.Interface().(string)
		r := regexp.MustCompile(`\(\(.*\)\)`)

		if r.MatchString(secret) {
			if secretTag != "true" {
				*errs = append(*errs, UnsupportedSecretError(fieldName))
				return
			}

			if platform.IsConcourse() {
				if !validateKeyValueSecret(secret) {
					*errs = append(*errs, InvalidSecretConcourseError(secret, fieldName))
					return
				}
			}

			if platform.IsActions() {
				splitOnSpaceSecret := strings.Split(secret, " ")
				if len(splitOnSpaceSecret) == 2 {
					if !validateSecretAbsolutePath(secret) {
						*errs = append(*errs, InvalidSecretActionsError(secret, fieldName))
						return
					}
				} else {
					if !validateKeyValueSecret(secret) {
						*errs = append(*errs, InvalidSecretActionsError(secret, fieldName))
						return
					}
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
			s.validate(success, fieldName, secretTag, errs, platform)
		}

		for ni, success := range notifications.OnFailure {
			fieldName := fmt.Sprintf("%s.on_failure[%d]", fieldName, ni)
			s.validate(success, fieldName, secretTag, errs, platform)
		}
		return

	default:
		panic(fmt.Sprintf("Not implemented for %s", v.Type()))
	}

}

func validateKeyValueSecret(secret string) bool {
	splitSecret := strings.Split(secret, ".")
	if len(splitSecret) != 2 || !regexp.MustCompile(`^\(\([a-zA-Z0-9\-_\.]+\)\)$`).MatchString(secret) {

		return false
	}

	return true
}

func validateSecretAbsolutePath(secret string) bool {
	splitSecret := strings.Split(secret, " ")

	prefix := strings.HasPrefix(splitSecret[0], "((/")
	containsSlash := strings.Contains(splitSecret[1], "/")

	return prefix && !containsSlash
}

func (s secretValidator) Validate(man Manifest) (errors []error) {
	var errs []error
	s.validate(man, "", "", &errs, man.Platform)
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
