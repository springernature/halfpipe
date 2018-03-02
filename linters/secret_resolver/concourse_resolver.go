package secret_resolver

import (
	"path"

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

	mapKey, secretKey := SecretToMapAndKey(concourseSecret)

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
