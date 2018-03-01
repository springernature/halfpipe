package secret_resolver

import (
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"errors"
)

const prefix = "prefix"
const team = "team"
const pipeline = "pipeline"
const mapKey = "mapKey"
const secretKey = "secretKey"
const concourseSecret = "((" + mapKey + "." + secretKey + "))"

type SecretResolverDouble struct {
	exists func(path, secretKey string) (bool, error)
}

func (y SecretResolverDouble) Exists(path string, secretKey string) (exists bool, err error) {
	return y.exists(path, secretKey)
}

func TestCallsOutToTwoDifferentPathsAndReturnsErrorIfNotFound(t *testing.T) {
	var calls [][]string
	resolver := NewConcourseResolver(prefix, SecretResolverDouble{
		exists: func(path string, secretKey string) (bool, error) {
			calls = append(calls, []string{path, secretKey})
			return false, nil
		},
	})

	err := resolver.Exists(team, pipeline, concourseSecret)

	assert.Len(t, calls, 2)
	assert.Contains(t, calls, []string{path.Join(prefix, team, pipeline, mapKey), secretKey})
	assert.Contains(t, calls, []string{path.Join(prefix, team, mapKey), secretKey})

	assert.NotNil(t, err)
}

func TestCallsOutOnlyOnceIfWeFindTheSecretInTheFirstPath(t *testing.T) {
	var calls [][]string
	resolver := NewConcourseResolver(prefix, SecretResolverDouble{
		exists: func(path string, secretKey string) (bool, error) {
			calls = append(calls, []string{path, secretKey})
			return true, nil
		},
	})

	err := resolver.Exists(team, pipeline, concourseSecret)

	assert.Len(t, calls, 1)
	assert.Contains(t, calls, []string{path.Join(prefix, team, pipeline, mapKey), secretKey})
	assert.Nil(t, err)
}

func TestCallsOutTwiceAndReturnsNilIfFoundInSecondCall(t *testing.T) {
	var calls [][]string
	resolver := NewConcourseResolver(prefix, SecretResolverDouble{
		exists: func(path string, secretKey string) (bool, error) {
			calls = append(calls, []string{path, secretKey})
			if len(calls) == 2 {
				return true, nil
			}
			return false, nil
		},
	})

	err := resolver.Exists(team, pipeline, concourseSecret)

	assert.Len(t, calls, 2)
	assert.Contains(t, calls, []string{path.Join(prefix, team, pipeline, mapKey), secretKey})
	assert.Contains(t, calls, []string{path.Join(prefix, team, mapKey), secretKey})
	assert.Nil(t, err)

}


func TestPassesOnTheError(t *testing.T) {
	myError := errors.New("Wryyy")
	resolver := NewConcourseResolver(prefix, SecretResolverDouble{
		exists: func(path string, secretKey string) (bool, error) {
			return false, myError
		},
	})

	err := resolver.Exists(team, pipeline, concourseSecret)
	assert.Equal(t, myError, err)

}

