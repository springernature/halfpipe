package mapper

import (
	"github.com/springernature/halfpipe/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

type testMapper struct {
	mapper func(original manifest.Manifest) (updated manifest.Manifest)
}

func (m testMapper) Apply(original manifest.Manifest) (updated manifest.Manifest) {
	return m.mapper(original)
}

func TestAppliesTheMappersCorrectly(t *testing.T) {
	var calledMapper1 bool
	returnManifestFrom1 := manifest.Manifest{Pipeline: "fromMapper1"}
	mapper1 := testMapper{func(original manifest.Manifest) (updated manifest.Manifest) {
		calledMapper1 = true
		return returnManifestFrom1
	}}

	var calledMapper2 bool
	var inputManifestToMapper2 manifest.Manifest
	returnManifestFrom2 := manifest.Manifest{Pipeline: "fromMapper2"}
	mapper2 := testMapper{mapper: func(original manifest.Manifest) (updated manifest.Manifest) {
		inputManifestToMapper2 = original
		calledMapper2 = true
		return returnManifestFrom2
	}}

	m := mapper{
		mappers: []Mapper{
			mapper1,
			mapper2,
		},
	}

	updated := m.Apply(manifest.Manifest{})

	assert.True(t, calledMapper1)

	assert.True(t, calledMapper2)
	assert.Equal(t, returnManifestFrom1, inputManifestToMapper2)

	assert.Equal(t, returnManifestFrom2, updated)

}
