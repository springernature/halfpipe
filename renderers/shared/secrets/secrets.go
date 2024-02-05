package secrets

import (
	"fmt"
	"strings"
)

// Secret models a Vault secret
// MapPath is root-relative path e.g. "/myteam/myproject/mysecretmap"
type Secret struct {
	MapPath string
	Key     string
}

// New returns a Secret from a string in the "halfpipe" format
// "((map.key))" or "((/path/to/map key))"
func New(s string, team string) *Secret {
	if !IsSecret(s) {
		return nil
	}

	secretValue := strings.TrimSpace(s[2 : len(s)-2])

	if isKeyValueSecret(secretValue) {
		parts := strings.Split(secretValue, ".")
		if isSharedSecret(parts[0]) {
			team = "shared"
		}
		return &Secret{
			MapPath: fmt.Sprintf("%s/%s", team, parts[0]),
			Key:     parts[1],
		}
	}

	if isAbsolutePathSecret(secretValue) {
		parts := strings.Split(secretValue, " ")
		mapPath := strings.TrimPrefix(parts[0], "/springernature/data/")
		mapPath = strings.TrimPrefix(mapPath, "/springernature/")
		return &Secret{
			MapPath: mapPath,
			Key:     parts[1],
		}
	}

	return nil
}

func IsSecret(s string) bool {
	return strings.HasPrefix(s, "((") && strings.HasSuffix(s, "))")
}

func isAbsolutePathSecret(s string) bool {
	return len(strings.Split(s, " ")) == 2
}

func isKeyValueSecret(s string) bool {
	return len(strings.Split(s, ".")) == 2
}

func isSharedSecret(s string) bool {
	return map[string]bool{
		"PPG-gradle-version-reporter":         true,
		"PPG-owasp-dependency-reporter":       true,
		"artifactory":                         true,
		"artifactory-support":                 true,
		"artifactory_test":                    true,
		"bla":                                 true,
		"burpsuiteenterprise":                 true,
		"content_hub-casper-credentials-live": true,
		"content_hub-casper-credentials-qa":   true,
		"contrastsecurity":                    true,
		"eas-sigrid":                          true,
		"ee-sso-route-service":                true,
		"fastly":                              true,
		"grafana":                             true,
		"halfpipe-artifacts":                  true,
		"halfpipe-docker-config":              true,
		"halfpipe-gcr":                        true,
		"halfpipe-github":                     true,
		"halfpipe-ml-deploy":                  true,
		"halfpipe-semver":                     true,
		"halfpipe-slack":                      true,
		"katee-tls-dev":                       true,
		"katee-tls-prod":                      true,
		"sentry-release-integration":          true,
	}[s]
}
