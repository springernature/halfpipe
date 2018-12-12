package retrigger_test

import (
	"github.com/springernature/halfpipe/retrigger"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetErrored(t *testing.T) {
	t.Run("Returns nothing if there are no builds", func(t *testing.T) {
		assert.Empty(t, retrigger.Builds{}.GetErrored())
	})

	t.Run("Returns nothing if there are no errored builds", func(t *testing.T) {
		assert.Empty(t, retrigger.Builds{
			{
				Status: "succeeded",
			},
			{
				Status: "failed",
			},
			{
				Status: "started",
			},
		}.GetErrored())
	})

	t.Run("Returns the errored builds", func(t *testing.T) {
		assert.Len(t, retrigger.Builds{
			{Status: "started"},
			{Status: "errored"},
			{Status: "failed"},
			{Status: "errored"},
			{Status: "aborted"},
			{Status: "pending"},
		}.GetErrored(), 2)

	})
}

func TestIsLatest(t *testing.T) {
	t.Run("Should return true when the errored build is the only build", func(t *testing.T) {
		build := retrigger.Build{
			Status: "errored",
		}
		builds := retrigger.Builds{build}

		assert.True(t, builds.IsLatest(build))
	})

	t.Run("Should return true when the errored build is the only build with that pipeline/job", func(t *testing.T) {
		erroredBuild := retrigger.Build{
			Status:       "errored",
			PipelineName: "pipeline",
			JobName:      "job",
		}
		builds := retrigger.Builds{
			erroredBuild,
			{PipelineName: "abc", JobName: "Kehe", Status: "errored"},
			{PipelineName: "qwerty", JobName: "qwerty", Status: "failed"},
		}
		assert.True(t, builds.IsLatest(erroredBuild))
	})

	t.Run("Should return true when the errored build is the latest build with that pipeline/job", func(t *testing.T) {
		erroredBuild := retrigger.Build{
			Status:       "errored",
			PipelineName: "pipeline",
			JobName:      "job",
			ID:           100,
		}
		successfulBuild := retrigger.Build{
			Status:       "successful",
			PipelineName: "pipeline",
			JobName:      "job",
			ID:           99,
		}
		builds := retrigger.Builds{
			erroredBuild,
			successfulBuild,
			{PipelineName: "abc", JobName: "Kehe", Status: "errored"},
			{PipelineName: "qwerty", JobName: "qwerty", Status: "failed"},
		}

		assert.True(t, builds.IsLatest(erroredBuild))
	})

	t.Run("Should return false when the errored build is not the latest build with that pipeline/job", func(t *testing.T) {
		successfulBuild := retrigger.Build{
			Status:       "successful",
			PipelineName: "pipeline",
			JobName:      "job",
			ID:           100,
		}

		erroredBuild := retrigger.Build{
			Status:       "errored",
			PipelineName: "pipeline",
			JobName:      "job",
			ID:           99,
		}

		builds := retrigger.Builds{
			successfulBuild,
			erroredBuild,
			{PipelineName: "abc", JobName: "Kehe", Status: "errored"},
			{PipelineName: "qwerty", JobName: "qwerty", Status: "failed"},
		}

		assert.False(t, builds.IsLatest(erroredBuild))
	})

	t.Run("Should do the right thing when there are multiple errored and successful builds", func(t *testing.T) {
		baseS := retrigger.Build{
			Status:       "successful",
			PipelineName: "pipeline",
			JobName:      "job",
		}
		baseF := baseS
		baseF.Status = "errored"

		s1 := baseS
		s1.ID = 1

		f2 := baseF
		f2.ID = 2

		s3 := baseS
		s3.ID = 3

		f4 := baseF
		f4.ID = 4

		builds := retrigger.Builds{s1, f2, s3, f4}

		assert.False(t, builds.IsLatest(s1))
		assert.False(t, builds.IsLatest(f2))
		assert.False(t, builds.IsLatest(s3))
		assert.True(t, builds.IsLatest(f4))
	})
}
