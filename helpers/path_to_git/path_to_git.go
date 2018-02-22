package path_to_git

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

func PathRelativeToGit(fs afero.Afero, startPath string, maxIterations int) (path string, err error) {
	return traverse(fs, startPath, maxIterations, []string{}, maxIterations)
}

func traverse(fs afero.Afero, currentPath string, maxIterations int, acc []string, ttl int) (path string, error error) {
	files, err := fs.ReadDir(currentPath)
	if err != nil {
		error = err
		return
	}

	for _, file := range files {
		if file.IsDir() && file.Name() == ".git" {
			paths := reverse(acc)
			path = strings.Join(paths, "/")
			return
		}
	}

	if ttl == 0 || currentPath == "/" {
		errStr := fmt.Sprintf("Could not find .git folder, traversed %d times up and ended up at %s", (maxIterations - ttl), currentPath)
		error = errors.New(errStr)
		return
	}

	parentPath := filepath.Join(currentPath, "../")
	return traverse(fs, parentPath, maxIterations, append(acc, filepath.Base(currentPath)), ttl-1)
}

func reverse(s []string) []string {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}
