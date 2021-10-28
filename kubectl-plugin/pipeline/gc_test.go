package pipeline

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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

func Test_getCompletionTimeFromObject(t *testing.T) {
	completionTime := time.Now()
	completionTimeStr := completionTime.Format(time.RFC3339)
	expectedCompletionTime, _ := time.Parse(time.RFC3339, completionTimeStr)
	type args struct {
		obj map[string]interface{}
	}
	tests := []struct {
		name               string
		args               args
		wantCompletionTime time.Time
		wantErr            bool
	}{{
		name: "Nil object",
		args: args{
			obj: nil,
		},
		wantErr: true,
	}, {
		name: "Without completionTime field",
		args: args{
			obj: map[string]interface{}{
				"status": map[string]interface{}{
					"updateTime": "",
				},
			},
		},
		wantErr: true,
	}, {
		name: "With an empty completionTime",
		args: args{
			obj: map[string]interface{}{
				"status": map[string]interface{}{
					"completionTime": "",
				},
			},
		},
		wantErr: true,
	}, {
		name: "Has completion time with RFC3339 layout",
		args: args{
			obj: map[string]interface{}{
				"status": map[string]interface{}{
					"completionTime": completionTime.Format(time.RFC3339),
				},
			},
		},
		wantCompletionTime: expectedCompletionTime,
	}, {
		name: "Has a invalid completion time",
		args: args{
			obj: map[string]interface{}{
				"status": map[string]interface{}{
					"completionTime": completionTime.Format(time.RFC1123),
				},
			},
		},
		wantErr: true,
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCompletionTime, err := getCompletionTimeFromObject(tt.args.obj)
			if (err != nil) != tt.wantErr {
				t.Errorf("getCompletionTimeFromObject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotCompletionTime, tt.wantCompletionTime) {
				t.Errorf("getCompletionTimeFromObject() = %v, want %v", gotCompletionTime, tt.wantCompletionTime)
			}
		})
	}
}
