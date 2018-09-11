package tasks

import (
	"github.com/springernature/halfpipe/linters/errors"
	"github.com/springernature/halfpipe/manifest"
)

func LintConsumerIntegrationTestTask(cit manifest.ConsumerIntegrationTest, providerHostRequired bool) (errs []error, warnings []error) {
	if cit.Consumer == "" {
		errs = append(errs, errors.NewMissingField("consumer"))
	}
	if cit.ConsumerHost == "" {
		errs = append(errs, errors.NewMissingField("consumer_host"))
	}
	if providerHostRequired {
		if cit.ProviderHost == "" {
			errs = append(errs, errors.NewMissingField("provider_host"))
		}
	}
	if cit.Script == "" {
		errs = append(errs, errors.NewMissingField("script"))
	}

	if cit.Retries < 0 || cit.Retries > 5 {
		errs = append(errs, errors.NewInvalidField("retries", "must be between 0 and 5"))
	}
	return
}
