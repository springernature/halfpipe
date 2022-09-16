package dependabot

import (
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestWalker(t *testing.T) {
	t.Run("It errors if not in git root", func(t *testing.T) {
		fs := afero.Afero{Fs: afero.NewMemMapFs()}
		_, err := NewWalker(fs, 3, []string{}).Walk()
		assert.Equal(t, ErrNotInGitRoot, err)
	})

	t.Run("It doesn't find .git", func(t *testing.T) {
		fs := afero.Afero{Fs: afero.NewMemMapFs()}
		fs.WriteFile(".git/blah", []byte(""), os.ModeAppend)
		fs.WriteFile("rootFile", []byte(""), os.ModeAppend)
		paths, err := NewWalker(fs, 3, []string{}).Walk()
		assert.NoError(t, err)
		assert.Equal(t, []string{"rootFile"}, paths)
	})

	t.Run("It stops scanning after depth and filters out folders", func(t *testing.T) {
		fs := afero.Afero{Fs: afero.NewMemMapFs()}
		fs.WriteFile(".git/blah", []byte(""), os.ModeAppend)
		fs.WriteFile("rootFile", []byte(""), os.ModeAppend)
		fs.WriteFile("1/1", []byte(""), os.ModeAppend)
		fs.WriteFile("1/2/2", []byte(""), os.ModeAppend)
		fs.WriteFile("1/2/3/3", []byte(""), os.ModeAppend)
		fs.WriteFile("1/2/3/4/4", []byte(""), os.ModeAppend)

		paths, err := NewWalker(fs, 0, []string{}).Walk()
		assert.NoError(t, err)
		assert.Equal(t, []string{"rootFile"}, paths)

		paths, err = NewWalker(fs, 1, []string{}).Walk()
		assert.NoError(t, err)
		assert.Equal(t, []string{"1/1", "rootFile"}, paths)

		paths, err = NewWalker(fs, 2, []string{}).Walk()
		assert.NoError(t, err)
		assert.Equal(t, []string{"1/1", "1/2/2", "rootFile"}, paths)

		paths, err = NewWalker(fs, 3, []string{}).Walk()
		assert.NoError(t, err)
		assert.Equal(t, []string{"1/1", "1/2/2", "1/2/3/3", "rootFile"}, paths)

		paths, err = NewWalker(fs, 4, []string{}).Walk()
		assert.NoError(t, err)
		assert.Equal(t, []string{"1/1", "1/2/2", "1/2/3/3", "1/2/3/4/4", "rootFile"}, paths)
	})

	t.Run("It filters out node_modules", func(t *testing.T) {
		fs := afero.Afero{Fs: afero.NewMemMapFs()}
		fs.WriteFile(".git/blah", []byte(""), os.ModeAppend)
		fs.WriteFile("rootFile", []byte(""), os.ModeAppend)

		fs.WriteFile("1/1", []byte(""), os.ModeAppend)
		fs.WriteFile("1/node_modules/1", []byte(""), os.ModeAppend)

		fs.WriteFile("1/2/2", []byte(""), os.ModeAppend)
		fs.WriteFile("1/2/node_modules/1", []byte(""), os.ModeAppend)

		fs.WriteFile("1/2/3/3", []byte(""), os.ModeAppend)
		fs.WriteFile("1/2/3/node_modules/1", []byte(""), os.ModeAppend)

		fs.WriteFile("1/2/3/4/4", []byte(""), os.ModeAppend)
		fs.WriteFile("1/2/3/4/node_modules/1", []byte(""), os.ModeAppend)

		paths, err := NewWalker(fs, 4, []string{}).Walk()
		assert.NoError(t, err)
		assert.Equal(t, []string{"1/1", "1/2/2", "1/2/3/3", "1/2/3/4/4", "rootFile"}, paths)
	})

	t.Run("It filters out skip_folders", func(t *testing.T) {
		fs := afero.Afero{Fs: afero.NewMemMapFs()}
		fs.WriteFile(".git/blah", []byte(""), os.ModeAppend)
		fs.WriteFile("rootFile", []byte(""), os.ModeAppend)

		fs.WriteFile("1/1", []byte(""), os.ModeAppend)
		fs.WriteFile("1/2/2", []byte(""), os.ModeAppend)
		fs.WriteFile("1/2/3/3", []byte(""), os.ModeAppend)
		fs.WriteFile("1/2/3/4/4", []byte(""), os.ModeAppend)

		fs.WriteFile("2/1", []byte(""), os.ModeAppend)
		fs.WriteFile("2/2/2", []byte(""), os.ModeAppend)
		fs.WriteFile("2/2/3/3", []byte(""), os.ModeAppend)
		fs.WriteFile("2/2/3/4/4", []byte(""), os.ModeAppend)

		fs.WriteFile("3/1", []byte(""), os.ModeAppend)
		fs.WriteFile("3/2/2", []byte(""), os.ModeAppend)
		fs.WriteFile("3/2/3/3", []byte(""), os.ModeAppend)
		fs.WriteFile("3/2/3/4/4", []byte(""), os.ModeAppend)

		paths, err := NewWalker(fs, 4, []string{"2", "3/2/"}).Walk()
		assert.NoError(t, err)
		assert.Equal(t, []string{"1/1", "1/2/2", "1/2/3/3", "1/2/3/4/4", "3/1", "rootFile"}, paths)
	})
}
