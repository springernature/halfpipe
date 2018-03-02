package manifest

import (
	"testing"

	"github.com/springernature/halfpipe/linters/errors"
	"github.com/stretchr/testify/assert"
)

func TestValidYaml(t *testing.T) {
	man, errs := Parse("team: my team")
	expected := Manifest{Team: "my team"}

	assert.Nil(t, errs)
	assert.Equal(t, expected, man)
}

func TestInvalidYaml(t *testing.T) {
	_, errs := Parse("team : { foo")

	assert.Equal(t, len(errs), 1)
}

func TestRepo(t *testing.T) {
	man, errs := Parse("repo: { uri: myuri, private_key: mypk }")
	expected := Manifest{
		Repo: Repo{
			Uri:        "myuri",
			PrivateKey: "mypk",
		},
	}

	assert.Nil(t, errs)
	assert.Equal(t, expected, man)
}

func TestRepoWithPaths(t *testing.T) {
	man, errs := Parse(`repo: { watched_paths: ["a", "b"] }`)
	expected := Manifest{
		Repo: Repo{
			WatchedPaths: []string{"a", "b"},
		},
	}

	assert.Nil(t, errs)
	assert.Equal(t, expected, man)

	///

	man, errs = Parse(`repo: { ignored_paths: ["a", "b"] }`)
	expected = Manifest{
		Repo: Repo{
			IgnoredPaths: []string{"a", "b"},
		},
	}

	assert.Nil(t, errs)
	assert.Equal(t, expected, man)

	///

	man, errs = Parse(`repo: { watched_paths: ["a", "b"], ignored_paths: ["c", "d"] }`)
	expected = Manifest{
		Repo: Repo{
			WatchedPaths: []string{"a", "b"},
			IgnoredPaths: []string{"c", "d"},
		},
	}

	assert.Nil(t, errs)
	assert.Equal(t, expected, man)
}

func TestRunTask(t *testing.T) {
	man, errs := Parse("tasks: [{ name: run, docker: {image: alpine}, script: build.sh, vars: { FOO: Foo, BAR: Bar } }]")
	expected := Manifest{
		Tasks: []Task{
			Run{
				Docker: Docker{
					Image: "alpine",
				},
				Script: "build.sh",
				Vars: Vars{
					"FOO": "Foo",
					"BAR": "Bar",
				},
			},
		},
	}

	assert.Nil(t, errs)
	assert.Equal(t, expected, man)
}

func TestMultipleTasks(t *testing.T) {
	man, errs := Parse("tasks: [{ name: run, docker: {image: img}, script: build.sh }, { name: docker-push, username: bob }, { name: run }, { name: deploy-cf, org: foo }]")
	expected := Manifest{
		Tasks: []Task{
			Run{
				Docker: Docker{
					Image: "img",
				},
				Script: "build.sh",
			},
			DockerPush{
				Username: "bob",
			},
			Run{},
			DeployCF{
				Org: "foo",
			},
		},
	}

	assert.Nil(t, errs)
	assert.Equal(t, expected, man)
}

func TestInvalidTask(t *testing.T) {
	_, errs := Parse("tasks: [{ name: unknown, foo: bar }]")

	assert.Equal(t, len(errs), 1)
}

func TestReportMultipleInvalidTasks(t *testing.T) {
	_, errs := Parse("tasks: [{ name: unknown, foo: bar }, { name: run, image: alpine, script: build.sh }, { notname: foo }]")

	assert.Equal(t, len(errs), 2)
	assert.IsType(t, errs[0], errors.NewInvalidField("", ""))
	assert.IsType(t, errs[1], errors.NewInvalidField("", ""))
}

func TestVarsParsedAsString(t *testing.T) {
	man, errs := Parse(`
tasks:
- name: run
  docker: 
    image: alpine
  script: build.sh
  vars:
    STRING: Foo Bar
    FLOAT: 4.2
    BOOL: true	
`)

	expected := Manifest{
		Tasks: []Task{
			Run{
				Docker: Docker{
					Image: "alpine",
				},
				Script: "build.sh",
				Vars: Vars{
					"STRING": "Foo Bar",
					"FLOAT":  "4.2",
					"BOOL":   "true",
				},
			},
		},
	}

	assert.Nil(t, errs)
	assert.Equal(t, expected, man)
}

func TestInvalidVars(t *testing.T) {
	_, errs := Parse(`
tasks:
- name: run
  image: alpine
  script: build.sh
  vars:
    EMPTY:
`)
	expected := errors.NewInvalidField("task", "")
	assert.IsType(t, expected, errs[0])
}

func TestSaveArtifact(t *testing.T) {
	manifest, errs := Parse(`
tasks:
- name: run
  image: alpine
  script: build.sh
  save_artifacts:
    - path/to/artifact.jar
`)

	assert.Nil(t, errs)
	assert.Equal(t, []string{"path/to/artifact.jar"}, manifest.Tasks[0].(Run).SaveArtifacts)

}

func TestDeployArtifact(t *testing.T) {
	manifest, errs := Parse(`
tasks:
- name: deploy-cf
  image: alpine
  script: build.sh
  deploy_artifact: path/to/artifact.jar
`)

	assert.Nil(t, errs)
	assert.Equal(t, "path/to/artifact.jar", manifest.Tasks[0].(DeployCF).DeployArtifact)

}
