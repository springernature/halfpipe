package manifest

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestName(t *testing.T) {
	pt := PipelineTrigger{
		Pipeline: "oscar-sites-bmc",
		Job:      "Deploy to QA-Preview (SNPaaS)",
	}

	assert.Equal(t, "oscar-sites-bmc.deploy-to-qa-preview-snpaas", pt.GetTriggerName())
}
