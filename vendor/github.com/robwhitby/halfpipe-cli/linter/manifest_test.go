package linter

import (
	"testing"

	. "github.com/robwhitby/halfpipe-cli/model"
	"github.com/stretchr/testify/assert"
)

var validRunTask = Run{
	Script: "./build.sh",
	Image:  "alpine",
}

var validManifest = Manifest{
	Team:  "ee",
	Repo:  Repo{Uri: "http://github.com/foo/bar.git"},
	Tasks: []Task{validRunTask},
}

func TestValidManifest(t *testing.T) {
	errs := LintManifest(validManifest)
	assert.Empty(t, errs)
}

func TestEmptyManifest(t *testing.T) {
	man := Manifest{}
	errs := LintManifest(man)

	assert.Len(t, errs, 3, "total 3 errs")
	assert.Contains(t, errs, NewMissingField("team"))
	assert.Contains(t, errs, NewMissingField("repo.uri"))
	assert.Contains(t, errs, NewMissingField("tasks"))
}

func TestRepo_UriFormat(t *testing.T) {
	man := validManifest
	man.Repo.Uri = "blah"
	errs := LintManifest(man)

	assert.Equal(t, errs[0], NewInvalidField("repo.uri", "must contain 'github'"))
}

func TestRunTask_Valid(t *testing.T) {
	var errs []error
	lintRunTask(validRunTask, 1, &errs)

	assert.Empty(t, errs)
}

func TestRunTask_MissingScript(t *testing.T) {
	run := validRunTask
	run.Script = ""

	var errs []error
	lintRunTask(run, 1, &errs)

	assert.Len(t, errs, 1)
	assert.Equal(t, NewMissingField("task 1: script"), errs[0])
}

func TestRunTask_MissingImage(t *testing.T) {
	run := validRunTask
	run.Image = ""

	var errs []error
	lintRunTask(run, 1, &errs)

	assert.Len(t, errs, 1)
	assert.Equal(t, NewMissingField("task 1: image"), errs[0])
}

func TestRunTask_Empty(t *testing.T) {
	run := Run{}

	var errs []error
	lintRunTask(run, 2, &errs)

	assert.Len(t, errs, 2)
	assert.Equal(t, NewMissingField("task 2: script"), errs[0])
	assert.Equal(t, NewMissingField("task 2: image"), errs[1])
}

type unknownTask struct {
	foo string
}

func (unknownTask) GetName() string {
	return "??"
}

func TestUnknownTask(t *testing.T) {
	man := validManifest
	man.Tasks = []Task{unknownTask{"foo"}}

	errs := LintManifest(man)

	assert.Len(t, errs, 1)
	assert.IsType(t, NewInvalidField("", ""), errs[0])
}
