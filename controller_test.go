package halfpipe

import (
	"errors"
	"github.com/concourse/concourse/atc"
	"github.com/springernature/halfpipe/linters"
	"github.com/springernature/halfpipe/linters/result"
	"github.com/springernature/halfpipe/manifest"
	"github.com/springernature/halfpipe/pipeline/actions"
	"github.com/stretchr/testify/assert"
	"testing"
)

type mockDefaulter struct {
	f func(original manifest.Manifest) manifest.Manifest
}

func (m mockDefaulter) Apply(original manifest.Manifest) (updated manifest.Manifest) {
	if m.f == nil {
		return original
	}
	return m.f(original)
}

type mockMapper struct {
	f func(original manifest.Manifest) (updated manifest.Manifest, err error)
}

func (m mockMapper) Apply(original manifest.Manifest) (updated manifest.Manifest, err error) {
	if m.f == nil {
		return original, nil
	}
	return m.f(original)
}

type mockConcourseRenderer struct {
	f func(manifest manifest.Manifest) atc.Config
}

func (m mockConcourseRenderer) Render(manifest manifest.Manifest) atc.Config {
	if m.f == nil {
		return atc.Config{}
	}
	return m.f(manifest)
}

type mockActionsRenderer struct {
	f func(manifest manifest.Manifest) actions.Actions
}

func (m mockActionsRenderer) Render(manifest manifest.Manifest) actions.Actions {
	if m.f == nil {
		return actions.Actions{}
	}
	return m.f(manifest)
}

type mockLinter struct {
	f func(manifest manifest.Manifest) result.LintResult
}

func (m mockLinter) Lint(manifest manifest.Manifest) result.LintResult {
	if m.f == nil {
		return result.LintResult{}
	}
	return m.f(manifest)
}

func TestController(t *testing.T) {
	var defaulterCalled bool
	var mapperCalled bool
	var concourseRendererCalled bool
	var linter1Called bool
	var linter2Called bool

	defaulter := mockDefaulter{f: func(original manifest.Manifest) manifest.Manifest {
		defaulterCalled = true
		return original
	}}

	mapper := mockMapper{f: func(original manifest.Manifest) (updated manifest.Manifest, err error) {
		mapperCalled = true
		return original, nil
	}}

	concouresRenderer := mockConcourseRenderer{f: func(manifest manifest.Manifest) atc.Config {
		concourseRendererCalled = true
		return atc.Config{}
	}}

	linter1 := mockLinter{f: func(manifest manifest.Manifest) result.LintResult {
		linter1Called = true
		return result.LintResult{}
	}}

	linter2 := mockLinter{f: func(manifest manifest.Manifest) result.LintResult {
		linter2Called = true
		return result.LintResult{}
	}}

	_, result := NewController(defaulter, mapper, []linters.Linter{linter1, linter2}, concouresRenderer, nil).Process(manifest.Manifest{})
	assert.True(t, defaulterCalled)
	assert.True(t, mapperCalled)
	assert.True(t, linter1Called)
	assert.True(t, linter2Called)
	assert.True(t, concourseRendererCalled)
	assert.False(t, result.HasErrors())
	assert.False(t, result.HasWarnings())
}

func TestControllerPassesOnLintErrors(t *testing.T) {
	var defaulterCalled bool
	var linter1Called bool
	var linter2Called bool

	defaulter := mockDefaulter{f: func(original manifest.Manifest) manifest.Manifest {
		defaulterCalled = true
		return original
	}}

	linter1 := mockLinter{f: func(manifest manifest.Manifest) result.LintResult {
		linter1Called = true
		return result.LintResult{}
	}}

	linter2 := mockLinter{f: func(manifest manifest.Manifest) result.LintResult {
		linter2Called = true
		return result.LintResult{
			Errors: []error{errors.New("ajajaj")},
		}
	}}

	_, result := NewController(defaulter, nil, []linters.Linter{linter1, linter2}, nil, nil).Process(manifest.Manifest{})
	assert.True(t, defaulterCalled)
	assert.True(t, linter1Called)
	assert.True(t, linter2Called)
	assert.True(t, result.HasErrors())
}

func TestControllerPassesOnMapperErrors(t *testing.T) {
	expectedErr := errors.New("blargh")
	mapper := mockMapper{f: func(original manifest.Manifest) (updated manifest.Manifest, err error) {
		return original, expectedErr
	}}

	_, result := NewController(mockDefaulter{}, mapper, nil, nil, nil).Process(manifest.Manifest{})
	assert.True(t, result.HasErrors())
	assert.Contains(t, result.Error(), expectedErr.Error())
}

func TestControllerCallsActionsRenderer(t *testing.T) {
	var defaulterCalled bool
	var mapperCalled bool
	var concourseRendererCalled bool
	var actionsRendererCalled bool

	defaulter := mockDefaulter{f: func(original manifest.Manifest) manifest.Manifest {
		defaulterCalled = true
		return original
	}}

	mapper := mockMapper{f: func(original manifest.Manifest) (updated manifest.Manifest, err error) {
		mapperCalled = true
		return original, nil
	}}

	concouresRenderer := mockConcourseRenderer{
		f: func(manifest manifest.Manifest) atc.Config {
			concourseRendererCalled = true
			return atc.Config{}
		},
	}

	actionsRenderer := mockActionsRenderer{f: func(manifest manifest.Manifest) actions.Actions {
		actionsRendererCalled = true
		return actions.Actions{}
	}}

	_, result := NewController(defaulter, mapper, nil, concouresRenderer, actionsRenderer).Process(manifest.Manifest{FeatureToggles: []string{manifest.FeatureGithubActions}})
	assert.True(t, defaulterCalled)
	assert.True(t, mapperCalled)
	assert.False(t, concourseRendererCalled)
	assert.True(t, actionsRendererCalled)
	assert.False(t, result.HasErrors())
	assert.False(t, result.HasWarnings())
}
