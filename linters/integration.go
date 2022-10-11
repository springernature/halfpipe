package linters

import (
	"github.com/springernature/halfpipe/manifest"
)

func LintConsumerIntegrationTestTask(cit manifest.ConsumerIntegrationTest, providerHostRequired bool) (errs []error) {
	if cit.Consumer == "" {
		errs = append(errs, NewErrMissingField("consumer"))
	}
	if cit.ConsumerHost == "" {
		errs = append(errs, NewErrMissingField("consumer_host"))
	}
	if providerHostRequired {
		if cit.ProviderHost == "" {
			errs = append(errs, NewErrMissingField("provider_host"))
		}
	}
	if cit.Script == "" {
		errs = append(errs, NewErrMissingField("script"))
	}

	if cit.Retries < 0 || cit.Retries > 5 {
		errs = append(errs, NewErrInvalidField("retries", "must be between 0 and 5"))
	}

	return errs
}
