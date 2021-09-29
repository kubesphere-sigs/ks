package pipeline

import (
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"testing"
)

func TestDescOrderWithCompletionTime(t *testing.T) {
	items := []unstructured.Unstructured{{
		Object: map[string]interface{}{
			"status": map[string]interface{}{
				"completionTime": "2021-09-28T01:39:13Z",
			},
			"flag": "0",
		},
	}, {
		Object: map[string]interface{}{
			"status": map[string]interface{}{
				"completionTime": "2021-09-30T01:39:13Z",
			},
			"flag": "1",
		},
	}, {
		Object: map[string]interface{}{
			"status": map[string]interface{}{
				"completionTime": "2021-09-27T01:39:13Z",
			},
			"flag": "2",
		},
	}}

	ascOrderWithCompletionTime(items)
	flag, _, _ := unstructured.NestedString(items[0].Object, "flag")
	assert.Equal(t, "2", flag)
}
