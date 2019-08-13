package parallel

import (
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWhenThereIsNothingToDo(t *testing.T) {
	input := manifest.TaskList{
		manifest.Run{},
		manifest.DockerCompose{},
		manifest.DockerPush{},
	}

	assert.Equal(t, input, NewParallelMerger().MergeParallelTasks(input))
}

func TestWithTrueParallelGroups(t *testing.T) {
	t.Run("with single parallels", func(t *testing.T) {
		input := manifest.TaskList{
			manifest.Run{},
			manifest.Run{Name: "p1", Parallel: "true"},
			manifest.Run{},
			manifest.Run{Name: "p2", Parallel: "true"},
		}

		expected := manifest.TaskList{
			manifest.Run{},
			manifest.Parallel{
				Tasks: manifest.TaskList{
					manifest.Run{Name: "p1"},
				},
			},
			manifest.Run{},
			manifest.Parallel{
				Tasks: manifest.TaskList{
					manifest.Run{Name: "p2"},
				},
			},
		}

		assert.Equal(t, expected, NewParallelMerger().MergeParallelTasks(input))
	})

	t.Run("with multiple subsequent parallels", func(t *testing.T) {
		input := manifest.TaskList{
			manifest.Run{},
			manifest.Run{Name: "t1-p", Parallel: "true"},
			manifest.Run{Name: "t2-p", Parallel: "true"},
			manifest.Run{},
			manifest.Run{Name: "t3-p", Parallel: "true"},
			manifest.Run{Name: "t3-p", Parallel: "true"},
		}

		expected := manifest.TaskList{
			manifest.Run{},
			manifest.Parallel{
				Tasks: manifest.TaskList{
					manifest.Run{Name: "t1-p"},
					manifest.Run{Name: "t2-p"},
				},
			},
			manifest.Run{},
			manifest.Parallel{
				Tasks: manifest.TaskList{
					manifest.Run{Name: "t3-p"},
					manifest.Run{Name: "t3-p"},
				},
			},
		}

		assert.Equal(t, expected, NewParallelMerger().MergeParallelTasks(input))
	})
}

func TestWithNamedParallelGroups(t *testing.T) {
	t.Run("with single parallels", func(t *testing.T) {
		input := manifest.TaskList{
			manifest.Run{Name: "p1", Parallel: "p1"},
			manifest.Run{},
			manifest.Run{},
			manifest.Run{Name: "p2", Parallel: "p2"},
		}

		expected := manifest.TaskList{
			manifest.Parallel{
				Tasks: manifest.TaskList{
					manifest.Run{Name: "p1"},
				},
			},
			manifest.Run{},
			manifest.Run{},
			manifest.Parallel{
				Tasks: manifest.TaskList{
					manifest.Run{Name: "p2"},
				},
			},
		}

		assert.Equal(t, expected, NewParallelMerger().MergeParallelTasks(input))
	})

	t.Run("with multiple subsequent parallels", func(t *testing.T) {
		input := manifest.TaskList{
			manifest.Run{},
			manifest.Run{Name: "t1", Parallel: "p1"},
			manifest.Run{Name: "t2", Parallel: "p1"},
			manifest.Run{},
			manifest.Run{Name: "t3", Parallel: "p2"},
			manifest.Run{Name: "t4", Parallel: "p2"},
			manifest.Run{},
		}

		expected := manifest.TaskList{
			manifest.Run{},
			manifest.Parallel{
				Tasks: manifest.TaskList{
					manifest.Run{Name: "t1"},
					manifest.Run{Name: "t2"},
				},
			},
			manifest.Run{},
			manifest.Parallel{
				Tasks: manifest.TaskList{
					manifest.Run{Name: "t3"},
					manifest.Run{Name: "t4"},
				},
			},
			manifest.Run{},
		}

		assert.Equal(t, expected, NewParallelMerger().MergeParallelTasks(input))
	})

	t.Run("with adjacent parallels", func(t *testing.T) {
		input := manifest.TaskList{
			manifest.Run{},
			manifest.Run{Name: "t1", Parallel: "p1"},
			manifest.Run{Name: "t2", Parallel: "p1"},
			manifest.Run{Name: "t3", Parallel: "p2"},
			manifest.Run{Name: "t4", Parallel: "p2"},
			manifest.Run{},
			manifest.Run{Name: "t5", Parallel: "p3"},
			manifest.Run{Name: "t6", Parallel: "p4"},
		}

		expected := manifest.TaskList{
			manifest.Run{},
			manifest.Parallel{
				Tasks: manifest.TaskList{
					manifest.Run{Name: "t1"},
					manifest.Run{Name: "t2"},
				},
			},
			manifest.Parallel{
				Tasks: manifest.TaskList{
					manifest.Run{Name: "t3"},
					manifest.Run{Name: "t4"},
				},
			},
			manifest.Run{},
			manifest.Parallel{
				Tasks: manifest.TaskList{
					manifest.Run{Name: "t5"},
				},
			},
			manifest.Parallel{
				Tasks: manifest.TaskList{
					manifest.Run{Name: "t6"},
				},
			},
		}

		assert.Equal(t, expected, NewParallelMerger().MergeParallelTasks(input))
	})
}

func TestWithMixedParallelGroups(t *testing.T) {
	input := manifest.TaskList{
		manifest.Run{Name: "t1"},
		manifest.Run{Name: "t2", Parallel: "p1"},
		manifest.Run{Name: "t3", Parallel: "p1"},
		manifest.Parallel{
			Tasks: manifest.TaskList{
				manifest.Run{Name: "t4", Parallel: "p2"},
				manifest.Run{Name: "t5", Parallel: "true"},
			},
		},
		manifest.Run{Name: "t6", Parallel: "p3"},
		manifest.Run{Name: "t7", Parallel: "p4"},
		manifest.Run{Name: "t8", Parallel: "true"},
		manifest.Run{Name: "t9", Parallel: "true"},
	}

	expected := manifest.TaskList{
		manifest.Run{Name: "t1"},
		manifest.Parallel{
			Tasks: manifest.TaskList{
				manifest.Run{Name: "t2"},
				manifest.Run{Name: "t3"},
			},
		},
		manifest.Parallel{
			Tasks: manifest.TaskList{
				manifest.Run{Name: "t4", Parallel: "p2"},
				manifest.Run{Name: "t5", Parallel: "true"},
			},
		},
		manifest.Parallel{
			Tasks: manifest.TaskList{
				manifest.Run{Name: "t6"},
			},
		},
		manifest.Parallel{
			Tasks: manifest.TaskList{
				manifest.Run{Name: "t7"},
			},
		},
		manifest.Parallel{
			Tasks: manifest.TaskList{
				manifest.Run{Name: "t8"},
				manifest.Run{Name: "t9"},
			},
		},
	}

	assert.Equal(t, expected, NewParallelMerger().MergeParallelTasks(input))
}
