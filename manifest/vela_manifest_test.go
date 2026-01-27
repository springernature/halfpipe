package manifest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshalVelaManifest(t *testing.T) {
	velaYaml := `apiVersion: core.oam.dev/v1beta1
kind: Application
metadata:
  name: my-app
  namespace: katee-my-team
spec:
  components:
    - name: my-component
      type: snstateless
      properties:
        image: nginx:latest
        env:
          - name: FOO
            value: bar
          - name: SECRET
            value: ${MY_SECRET}
`

	vm, err := UnmarshalVelaManifest([]byte(velaYaml))

	assert.NoError(t, err)
	assert.Equal(t, "Application", vm.Kind)
	assert.Equal(t, "my-app", vm.Metadata.Name)
	assert.Equal(t, "katee-my-team", vm.Metadata.Namespace)
}

func TestUnmarshalVelaManifest_InvalidYAML(t *testing.T) {
	invalidYaml := `this is not: valid: yaml:`

	_, err := UnmarshalVelaManifest([]byte(invalidYaml))

	assert.Error(t, err)
}
