package secret_resolver

import (
	"path"

	"strings"

	"github.com/springernature/halfpipe/linters/errors"
)

type ConcourseResolver interface {
	Exists(team string, pipeline string, concourseSecret string) (err error)
}

type concourseResolver struct {
	prefix          string
	secretsResolver SecretResolver
}

func NewConcourseResolver(prefix string, secretsResolver SecretResolver) concourseResolver {
	return concourseResolver{
		prefix:          prefix,
		secretsResolver: secretsResolver,
	}
}

func (c concourseResolver) Exists(team string, pipeline string, concourseSecret string) (err error) {

	mapKey, secretKey := c.secretToMapAndKey(concourseSecret)

	paths := []string{
		path.Join(c.prefix, team, pipeline, mapKey),
		path.Join(c.prefix, team, mapKey),
	}

	for _, p := range paths {
		exists, e := c.secretsResolver.Exists(p, secretKey)
		if e != nil {
			err = e
			return
		}
		if exists {
			return nil
		}
	}

	return errors.NewVaultSecretNotFoundError(c.prefix, team, pipeline, concourseSecret)
}

func (concourseResolver) secretToMapAndKey(secret string) (string, string) {
	s := strings.Replace(strings.Replace(secret, "((", "", -1), "))", "", -1)
	parts := strings.Split(s, ".")
	mapName, keyName := parts[0], parts[1]
	return mapName, keyName
}
