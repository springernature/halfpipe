package cmds

import (
	"github.com/springernature/halfpipe/config"
	"log"
	"os"
	"path"
	"path/filepath"
	"testing"
)

func findE2EPaths() []string {
	startingDir, _ := os.Getwd()

	var e2eTestPaths []string
	filepath.Walk("../../e2e/", func(filePath string, info os.FileInfo, err error) error {
		if info.IsDir() {
			fullPath := path.Join(startingDir, filePath)
			if _, err := os.Stat(path.Join(fullPath, ".halfpipe.io")); err == nil {
				if _, err := os.Stat(path.Join(fullPath, "test.sh")); err != nil {
					e2eTestPaths = append(e2eTestPaths, fullPath)
				}
			}
		}
		return nil
	})
	return e2eTestPaths
}

// does not fail if test output does not match expected pipeline
// only useful for checking code coverage of e2e tests
func TestE2EForCoverage(t *testing.T) {
	if os.Getenv("HALFPIPE_SKIP_COVERAGE_TESTS") == "true" {
		t.Skip("skipping test; $HALFPIPE_SKIP_COVERAGE_TESTS==true")
	}
	defer quiet()()
	config.CheckBranch = "false"
	for _, testPath := range findE2EPaths() {
		t.Run(testPath, func(t *testing.T) {
			os.Chdir(testPath)
			rootCmd.Run(nil, []string{})
		})
	}
}

func quiet() func() {
	null, _ := os.Open(os.DevNull)
	sout := os.Stdout
	serr := os.Stderr
	os.Stdout = null
	os.Stderr = null
	log.SetOutput(null)
	return func() {
		defer null.Close()
		os.Stdout = sout
		os.Stderr = serr
		log.SetOutput(os.Stderr)
	}
}
