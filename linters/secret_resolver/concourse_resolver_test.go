package secret_resolver

import (
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

const prefix = "prefix"
const team = "team"
const pipeline = "pipeline"
const mapKey = "mapKey"
const secretKey = "secretKey"
const concourseSecret = "((" + mapKey + "." + secretKey + "))"

type MockSecretsResolver struct {
}

var calls [][]string
var existsPath string

func (MockSecretsResolver) Exists(path string, secretKey string) (exists bool, err error) {
	calls = append(calls, []string{path, secretKey})

	if path == existsPath {
		return true, nil
	}
	return
}

func newConcourseResolver() ConcourseResolver {
	calls = [][]string{}
	existsPath = ""
	return NewConcourseResolver(prefix, MockSecretsResolver{})
}

func TestCallsOutToTwoDifferentPathsAndReturnsErrorIfNotFound(t *testing.T) {
	err := newConcourseResolver().Exists(team, pipeline, concourseSecret)

	path1 := path.Join(prefix, team, pipeline, mapKey)
	path2 := path.Join(prefix, team, mapKey)

	assert.Len(t, calls, 2)
	assert.Contains(t, calls, []string{path1, secretKey})
	assert.Contains(t, calls, []string{path2, secretKey})

	assert.NotNil(t, err)
}

func TestCallsOutOnlyOnceIfWeFindTheSecretInTheFirstPath(t *testing.T) {
	resolver := newConcourseResolver()

	path1 := path.Join(prefix, team, pipeline, mapKey)
	existsPath = path1

	err := resolver.Exists(team, pipeline, concourseSecret)

	assert.Len(t, calls, 1)
	assert.Contains(t, calls, []string{path1, secretKey})
	assert.Nil(t, err)
}
