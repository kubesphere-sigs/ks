package pipeline

import (
	"fmt"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/pipeline/option"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestDescOrderWithCompletionTime(t *testing.T) {
	items := []unstructured.Unstructured{{
		Object: map[string]interface{}{
			"flag": "3",
		},
	}, {
		Object: map[string]interface{}{
			"status": map[string]interface{}{
				"completionTime": "2021-09-28T01:39:13Z",
			},
			"flag": "0",
		},
	}, {
		Object: map[string]interface{}{
			"flag": "4",
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
	// make sure that object without status.completionTime be at the end of the result
	flag, _, _ = unstructured.NestedString(items[3].Object, "flag")
	assert.Equal(t, "3", flag)
	flag, _, _ = unstructured.NestedString(items[4].Object, "flag")
	assert.Equal(t, "4", flag)
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

// get pipelinerunList need to delete
func TestNeedToDelete(t *testing.T) {
	pipeline := &gcPipeline{
		option:          nil,
		name:            "test-pipeline",
		namespace:       "test-ns",
		daysToKeep:      7,
		numToKeep:       5,
		pipelinerunList: nil,
	}
	now := time.Now()

	// case1: days_to_keep
	pipeline.pipelinerunList = make([]*gcPipelinerun, 4)
	pipeline.daysToKeep = 2
	pipeline.pipelinerunList[0] = &gcPipelinerun{
		name:           "run-0",
		phase:          option.PipelinerunPhaseSucceeded,
		completionTime: now.AddDate(0, 0, -3),
	}
	pipeline.pipelinerunList[1] = &gcPipelinerun{
		name:           "run-1",
		phase:          option.PipelinerunPhaseSucceeded,
		completionTime: now.AddDate(0, 0, -2),
	}
	pipeline.pipelinerunList[2] = &gcPipelinerun{
		name:           "run-2",
		phase:          option.PipelinerunPhaseSucceeded,
		completionTime: now.AddDate(0, 0, -2),
	}
	pipeline.pipelinerunList[3] = &gcPipelinerun{
		name:           "run-3",
		phase:          option.PipelinerunPhaseSucceeded,
		completionTime: now.AddDate(0, 0, -1),
	}
	pipeline.pipelinerunList[2].completionTime = pipeline.pipelinerunList[2].completionTime.Add(10 * time.Minute)
	deletingList := pipeline.needToDelete()
	assert.EqualValues(t, []string{"run-0", "run-1"}, deletingList)

	// case2: num_to_keep
	pipeline.numToKeep = 1
	deletingList = pipeline.needToDelete()
	assert.EqualValues(t, []string{"run-0", "run-1", "run-2"}, deletingList)

	// case3: complex with un-completion and lastStable and lastSuccessful
	// pipelinerunList: {
	//   run0:-7/success, run1:-6/success, run2:-5/running, run3:-4/success
	//   run4:-3/failed, run5:-2/failed, run6:-1/failed, run7:0/running
	// }
	pipeline.pipelinerunList = make([]*gcPipelinerun, 8)
	pipeline.daysToKeep = 7
	pipeline.numToKeep = 5
	for i := 0; i < 8; i++ {
		pipeline.pipelinerunList[i] = &gcPipelinerun{
			name:           fmt.Sprintf("run-%d", i),
			phase:          option.PipelinerunPhaseSucceeded,
			completionTime: now.AddDate(0, 0, -7+i),
		}
	}
	ZeroTime, _ := time.ParseInLocation(time.RFC3339, "0000-00-00T00:00:00Z00:00", time.Local)
	pipeline.pipelinerunList[2].completionTime = ZeroTime // setup un-completion
	pipeline.pipelinerunList[2].phase = option.PipelinerunPhaseRunning
	pipeline.pipelinerunList[7].completionTime = ZeroTime // setup un-completion
	pipeline.pipelinerunList[7].phase = option.PipelinerunPhaseRunning
	pipeline.pipelinerunList[4].phase = option.PipelinerunPhaseFailed // setup Phase to failed
	pipeline.pipelinerunList[5].phase = option.PipelinerunPhaseFailed // setup Phase to failed
	pipeline.pipelinerunList[6].phase = option.PipelinerunPhaseFailed // setup Phase to failed
	deletingList = pipeline.needToDelete()
	assert.EqualValues(t, []string{"run-0", "run-1", "run-4"}, deletingList)

	pipeline.daysToKeep = 2
	deletingList = pipeline.needToDelete()
	assert.EqualValues(t, []string{"run-0", "run-1", "run-4", "run-5"}, deletingList)
}
