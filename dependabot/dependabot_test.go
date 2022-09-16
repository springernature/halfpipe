package dependabot

import (
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

type mockWalker struct {
	files                 []string
	err                   error
	calledWithDepth       int
	calledWithSkipFolders []string
}

func (m *mockWalker) Walk(depth int, skipFolders []string) ([]string, error) {
	m.calledWithDepth = depth
	m.calledWithSkipFolders = skipFolders
	return m.files, m.err
}

type mockFilter struct {
	calledWithFiles          []string
	calledWithSkipEcosystems []string
	called                   bool
}

func (m *mockFilter) Filter(paths []string, skipEcosystems []string) []string {
	m.calledWithFiles = paths
	m.calledWithSkipEcosystems = skipEcosystems
	m.called = true
	return paths
}

type mockRender struct {
	config          Config
	calledWithFiles []string
}

func (m *mockRender) Render(paths []string) Config {
	m.calledWithFiles = paths
	return m.config
}

func TestDependabot(t *testing.T) {
	t.Run("It returns error if walker returns error", func(t *testing.T) {
		expectedErr := errors.New("Expected")
		expectedDepth := 1337
		expectedSkipFolders := []string{"a", "b", "c"}
		w := mockWalker{err: expectedErr}
		f := mockFilter{}
		r := mockRender{}
		_, err := New(DependabotConfig{Depth: expectedDepth, SkipFolders: expectedSkipFolders}, &w, &f, &r).Resolve()
		assert.Equal(t, expectedErr, err)
		assert.Equal(t, expectedDepth, w.calledWithDepth)
		assert.Equal(t, expectedSkipFolders, w.calledWithSkipFolders)
		assert.False(t, f.called)
	})

	t.Run("It filters found paths", func(t *testing.T) {
		expectedFiles := []string{"file1", "file2", "file3"}
		expectedSkippedEcosystems := []string{"a", "b", "c"}
		expectedConfig := Config{Version: 2}

		w := mockWalker{files: expectedFiles}
		f := mockFilter{}
		r := mockRender{config: expectedConfig}

		config, err := New(DependabotConfig{SkipEcosystem: expectedSkippedEcosystems}, &w, &f, &r).Resolve()
		assert.NoError(t, err)
		assert.True(t, f.called)
		assert.Equal(t, expectedFiles, f.calledWithFiles)
		assert.Equal(t, expectedSkippedEcosystems, f.calledWithSkipEcosystems)
		assert.Equal(t, expectedFiles, r.calledWithFiles)
		assert.Equal(t, config, expectedConfig)
	})
}
