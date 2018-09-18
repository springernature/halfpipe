package upload

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSecurityQuestionWithWrongAnswer(t *testing.T) {
	stdout := new(bytes.Buffer)
	stdin := new(bytes.Buffer)

	cmd := SecurityQuestion("pipeline", "myFeature")
	err := cmd.Executor(stdout, stdin)
	assert.Equal(t, ErrWrongAnswer, err)

	assert.Contains(t, stdout.String(), "* You are on branch myFeature")
	assert.Contains(t, stdout.String(), "* We will upload the pipeline as pipeline-myFeature")
}

func TestSecurityQuestionWithCorrectAnswer(t *testing.T) {
	stdout := new(bytes.Buffer)
	stdin := new(bytes.Buffer)
	stdin.Write([]byte("y"))

	cmd := SecurityQuestion("pipeline", "myFeature")
	err := cmd.Executor(stdout, stdin)
	assert.Nil(t, err)
}
