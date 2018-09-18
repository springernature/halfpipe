package upload

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"bytes"
)

func TestSecurityQuestionWithWrongAnswer(t *testing.T) {
	stdout := new(bytes.Buffer)
	stdin := new(bytes.Buffer)

	cmd := SecurityQuestion(stdout, stdin, "pipeline", "myFeature")
	err := cmd.Executor()
	assert.Equal(t, ErrWrongAnswer, err)

	assert.Contains(t, stdout.String(), "* You are on branch myFeature")
	assert.Contains(t, stdout.String(), "* We will upload the pipeline as pipeline-myFeature")
}

func TestSecurityQuestionWithCorrectAnswer(t *testing.T) {
	stdout := new(bytes.Buffer)
	stdin := new(bytes.Buffer)
	stdin.Write([]byte("y"))

	cmd := SecurityQuestion(stdout, stdin, "pipeline", "myFeature")
	err := cmd.Executor()
	assert.Nil(t, err)
}
