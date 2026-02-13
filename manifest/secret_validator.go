package manifest

import (
	"code.cloudfoundry.org/cli/util/manifestparser"
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"strings"
)

const tagName = "secretAllowed"

var reservedKeyNames = []string{"value"}

var UnsupportedSecretError = func(fieldName string) error {
	return fmt.Errorf("'%s' is not allowed to contain a secret", fieldName)
}

var InvalidSecretConcourseError = func(secret, fieldName string) error {
	return fmt.Errorf("'%s' at '%s' is not a valid key, must be in format of ((mapName.keyName)) or ((path/to/mapName.keyName))", secret, fieldName)
}

var InvalidSecretActionsError = func(secret, fieldName string) error {
	return fmt.Errorf("'%s' at '%s' is not a valid key, must be in format of ((mapName.keyName)), ((path/to/mapName.keyName)) or ((/springernature/data/path/to/mapName keyName))", secret, fieldName)
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

func (s secretValidator) validate(i any, fieldName string, secretTag string, errs *[]error, platform Platform) {
	v := reflect.ValueOf(i)

	switch v.Type() {

	case reflect.TypeFor[Manifest](),
		reflect.TypeFor[Repo](),
		reflect.TypeFor[Run](),
		reflect.TypeFor[Docker](),
		reflect.TypeFor[DockerPush](),
		reflect.TypeFor[DockerPushAWS](),
		reflect.TypeFor[DockerCompose](),
		reflect.TypeFor[DeployCF](),
		reflect.TypeFor[DeployKatee](),
		reflect.TypeFor[ConsumerIntegrationTest](),
		reflect.TypeFor[DeployMLZip](),
		reflect.TypeFor[DeployMLModules](),
		reflect.TypeFor[ArtifactConfig](),
		reflect.TypeFor[GitTrigger](),
		reflect.TypeFor[TimerTrigger](),
		reflect.TypeFor[DockerTrigger](),
		reflect.TypeFor[Buildpack](),
		reflect.TypeFor[PipelineTrigger]():

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

	case reflect.TypeFor[TaskList]():
		for i, elem := range v.Interface().(TaskList) {
			realFieldName := fmt.Sprintf("%s[%d]", fieldName, i)
			s.validate(elem, realFieldName, secretTag, errs, platform)
		}
	case reflect.TypeFor[TriggerList]():
		for i, elem := range v.Interface().(TriggerList) {
			realFieldName := fmt.Sprintf("%s[%d]", fieldName, i)
			s.validate(elem, realFieldName, secretTag, errs, platform)
		}
	case reflect.TypeFor[Parallel]():
		for i, elem := range v.Interface().(Parallel).Tasks {
			realFieldName := fmt.Sprintf("%s[%d]", fieldName, i)
			s.validate(elem, realFieldName, secretTag, errs, platform)
		}
	case reflect.TypeFor[Sequence]():
		for i, elem := range v.Interface().(Sequence).Tasks {
			realFieldName := fmt.Sprintf("%s[%d]", fieldName, i)
			s.validate(elem, realFieldName, secretTag, errs, platform)
		}
	case reflect.TypeFor[[]string]():
		for i, elem := range v.Interface().([]string) {
			realFieldName := fmt.Sprintf("%s[%d]", fieldName, i)
			s.validate(elem, realFieldName, secretTag, errs, platform)
		}

	case reflect.TypeFor[FeatureToggles]():
		for i, elem := range v.Interface().(FeatureToggles) {
			realFieldName := fmt.Sprintf("%s[%d]", fieldName, i)
			s.validate(elem, realFieldName, secretTag, errs, platform)
		}

	case reflect.TypeFor[Vars]():
		for key, value := range v.Interface().(Vars) {
			realKeyName := fmt.Sprintf("key %s[%s]", fieldName, key)
			s.validate(key, realKeyName, "false", errs, "")
			realValueName := fmt.Sprintf("%s[%s]", fieldName, key)
			s.validate(value, realValueName, secretTag, errs, platform)
		}

	case reflect.TypeFor[string]():
		secret := v.Interface().(string)
		r := regexp.MustCompile(`\(\(.*\)\)`)

		if r.MatchString(secret) {
			if secretTag != "true" {
				*errs = append(*errs, UnsupportedSecretError(fieldName))
				return
			}

			if platform.IsConcourse() {
				if !validateKeyValueSecret(secret) && !validateMultipleLevelSecret(secret) {
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
					if !validateKeyValueSecret(secret) && !validateMultipleLevelSecret(secret) {
						*errs = append(*errs, InvalidSecretActionsError(secret, fieldName))
						return
					}
				}
			}
		}

	case reflect.TypeFor[Notifications]():
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

	// Stuff that we don't care about as they cannot contain secrets.
	case reflect.TypeFor[bool](),
		reflect.TypeFor[int](),
		reflect.TypeFor[manifestparser.Application](),
		reflect.TypeFor[Update](),
		reflect.TypeFor[Platform](),
		reflect.TypeFor[ComposeFiles](),
		reflect.TypeFor[GitHubEnvironment](),
		reflect.TypeFor[VelaManifest]():
		return

	default:
		panic(fmt.Sprintf("Not implemented for %s", v.Type()))
	}

}

func validateMultipleLevelSecret(secret string) bool {
	// regex matches ((path/to/secret.value)) and ((path/to/more/levels/secret.value))
	return regexp.MustCompile(`\(\(([a-zA-Z0-9\-_]+\/){1,}[a-zA-Z0-9\-_]+\.[a-zA-Z0-9\-_]+\)\)`).MatchString(secret)
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

	prefix := strings.HasPrefix(splitSecret[0], "((/springernature/data/")
	containsSlash := strings.Contains(splitSecret[1], "/")

	return prefix && !containsSlash
}

func (s secretValidator) Validate(man Manifest) (errors []error) {
	var errs []error
	s.validate(man, "", "", &errs, man.Platform)
	return errs
}

func (s secretValidator) IsReservedKeyName(keyName string) bool {
	return slices.ContainsFunc(reservedKeyNames, func(s string) bool { return s == keyName })
}
