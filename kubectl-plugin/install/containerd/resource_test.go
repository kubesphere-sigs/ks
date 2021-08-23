package containerd

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEmebdContent(t *testing.T) {
	assert.NotEmpty(t, GetCrictl(), "crictl config file is empty")
	assert.NotEmpty(t, GetConfigToml(), "containerd config file is empty")
}
