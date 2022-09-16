package dependabot

import (
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

type mockWalker struct {
	files []string
	err   error
}

func (m *mockWalker) Walk() ([]string, error) {
	return m.files, m.err
}

type mockFilter struct {
	calledWithFiles []string
	called          bool
}

func (m *mockFilter) Filter(paths []string) MatchedPaths {
	m.calledWithFiles = paths
	m.called = true
	return MatchedPaths{}
}

type mockRender struct {
	config          Config
	calledWithPaths MatchedPaths
}

func (m *mockRender) Render(paths MatchedPaths) Config {
	m.calledWithPaths = paths
	return m.config
}

func TestDependabot(t *testing.T) {
	t.Run("It returns error if walker returns error", func(t *testing.T) {
		expectedErr := errors.New("Expected")
		w := mockWalker{err: expectedErr}
		f := mockFilter{}
		r := mockRender{}
		_, err := New(&w, &f, &r).Resolve()
		assert.Equal(t, expectedErr, err)
		assert.False(t, f.called)
	})

	t.Run("It filters found paths", func(t *testing.T) {
		expectedFiles := []string{"file1", "file2", "file3"}
		expectedConfig := Config{Version: 2}

		w := mockWalker{files: expectedFiles}
		f := mockFilter{}
		r := mockRender{config: expectedConfig}

		config, err := New(&w, &f, &r).Resolve()
		assert.NoError(t, err)
		assert.True(t, f.called)
		assert.Equal(t, expectedFiles, f.calledWithFiles)
		assert.Equal(t, config, expectedConfig)
	})
}
