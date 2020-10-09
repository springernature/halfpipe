package actions_test

import (
	"fmt"
	"github.com/springernature/halfpipe/manifest"
	"github.com/springernature/halfpipe/renderers/actions"
	"testing"
)

func TestActions(t *testing.T) {

	halfpipeYaml := `
team: halfpipe-team
pipeline: halfpipe-e2e-run

triggers:
- type: git
  watched_paths:
  - e2e/actions/run

tasks:
- type: run
  name: This is a test
  script: ./a
  docker:
    image: alpine:test`

	man, _ := manifest.Parse(halfpipeYaml)
	workflow, _ := actions.NewActions().Render(man)

	fmt.Println(workflow)

}
