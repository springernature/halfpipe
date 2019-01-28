package pipeline

import (
	"github.com/concourse/atc"
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTimeout(t *testing.T) {

	task1 := "task1"
	task2 := "task2"
	task3 := "task3"
	task4 := "task4"
	task5 := "task5"

	shortTimeout := "5m"
	longTimeout := "5h"

	man := manifest.Manifest{
		Tasks: manifest.TaskList{
			manifest.Run{Name: task1, Timeout: shortTimeout},
			manifest.DeployCF{Name: task2, Timeout: longTimeout},
			manifest.DockerPush{Name: task3, Timeout: shortTimeout},
			manifest.ConsumerIntegrationTest{Name: task4, Timeout: shortTimeout},
			manifest.DeployMLModules{Name: task5, Timeout: longTimeout},
		},
	}

	config := testPipeline().Render(man)

	hasCorrectTimeout := func(jc atc.JobConfig, t string) bool {
		for _, p := range jc.Plan {
			if p.Timeout != t {
				return false
			}
		}
		return true
	}

	c, _ := config.Jobs.Lookup(task1)
	assert.True(t, hasCorrectTimeout(c, shortTimeout))

	c, _ = config.Jobs.Lookup(task2)
	assert.True(t, hasCorrectTimeout(c, longTimeout))

	c, _ = config.Jobs.Lookup(task3)
	assert.True(t, hasCorrectTimeout(c, shortTimeout))

	c, _ = config.Jobs.Lookup(task4)
	assert.True(t, hasCorrectTimeout(c, shortTimeout))

	c, _ = config.Jobs.Lookup(task5)
	assert.True(t, hasCorrectTimeout(c, longTimeout))
}

func TestTimeout2(t *testing.T) {

	task1 := "task1"
	task2 := "task2"
	task3 := "task3"
	task4 := "task4"
	task5 := "task5"

	shortTimeout := "5m"
	longTimeout := "5h"

	man := manifest.Manifest{
		Tasks: manifest.TaskList{
			manifest.Run{Name: task1, Timeout: shortTimeout},
			manifest.DeployCF{
				Name:    task2,
				Timeout: longTimeout,
				PrePromote: manifest.TaskList{
					manifest.Run{
						Timeout: "100h",
					},
				},
			},
			manifest.DockerPush{Name: task3, Timeout: shortTimeout},
			manifest.ConsumerIntegrationTest{Name: task4, Timeout: shortTimeout},
			manifest.DeployMLModules{Name: task5, Timeout: longTimeout},
		},
	}

	config := testPipeline().Render(man)

	hasCorrectTimeout := func(jc atc.JobConfig, t string) bool {
		for _, p := range jc.Plan {
			if p.Timeout != t {
				return false
			}
		}
		return true
	}

	c, _ := config.Jobs.Lookup(task1)
	assert.True(t, hasCorrectTimeout(c, shortTimeout))

	c, _ = config.Jobs.Lookup(task2)
	assert.True(t, hasCorrectTimeout(c, longTimeout))

	c, _ = config.Jobs.Lookup(task3)
	assert.True(t, hasCorrectTimeout(c, shortTimeout))

	c, _ = config.Jobs.Lookup(task4)
	assert.True(t, hasCorrectTimeout(c, shortTimeout))

	c, _ = config.Jobs.Lookup(task5)
	assert.True(t, hasCorrectTimeout(c, longTimeout))
}
