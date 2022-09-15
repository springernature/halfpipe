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

func TestDependabot(t *testing.T) {
	t.Run("It returns error if walker returns error", func(t *testing.T) {
		expectedErr := errors.New("Expected")
		expectedDepth := 1337
		expectedSkipFolders := []string{"a", "b", "c"}
		w := mockWalker{err: expectedErr}
		_, err := New(DependabotConfig{Depth: expectedDepth, SkipFolders: expectedSkipFolders}, &w).Resolve()
		assert.Equal(t, expectedErr, err)
		assert.Equal(t, expectedDepth, w.calledWithDepth)
		assert.Equal(t, expectedSkipFolders, w.calledWithSkipFolders)
	})
}
