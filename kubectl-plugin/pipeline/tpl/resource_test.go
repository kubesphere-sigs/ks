package tpl_test

import (
	"github.com/kubesphere-sigs/ks/kubectl-plugin/pipeline/tpl"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEmebdContent(t *testing.T) {
	assert.NotEmpty(t, tpl.GetLongRunPipeline(), "long run Pipeline content is empty")
	assert.NotEmpty(t, tpl.GetBuildJava(), "java building Pipeline content is empty")
	assert.NotEmpty(t, tpl.GetBuildGo(), "go building Pipeline content is empty")
	assert.NotEmpty(t, tpl.GetSimple(), "simple Pipeline content is empty")
}
