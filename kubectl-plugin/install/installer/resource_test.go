package installer

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEmebdContent(t *testing.T) {
	assert.NotEmpty(t, GetKSInstaller(), "ks installer file is empty")
	assert.NotEmpty(t, GetClusterConfig(), "cluster configuratioin config file is empty")
}
