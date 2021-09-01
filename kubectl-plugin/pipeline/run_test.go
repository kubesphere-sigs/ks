package pipeline

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"html/template"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
	"testing"
)

func TestPipelineRunTplParse(t *testing.T) {
	tpl := template.New("PipelineRunTpl")
	tpl, err := tpl.Parse(pipelineRunTpl)
	if err != nil {
		t.Errorf("failed to parse PipelineRun template, err = %v", err)
	}
	var buf bytes.Buffer
	err = tpl.Execute(&buf, map[string]interface{}{
		"name":       "fake_name",
		"namespace":  "fake_ns",
		"parameters": nil,
	})
	if err != nil {
		t.Errorf("failed to execute PipelineRun template, err = %v", err)
	}
	fmt.Println(buf.String())
}

// getNestedString comes from k8s.io/apimachinery@v0.19.4/pkg/apis/meta/v1/unstructured/helpers.go:277
func getNestedString(obj map[string]interface{}, fields ...string) string {
	val, found, err := unstructured.NestedString(obj, fields...)
	if !found || err != nil {
		return ""
	}
	return val
}

func getNestSlice(obj map[string]interface{}, fields ...string) []interface{} {
	val, found, err := unstructured.NestedSlice(obj, fields...)
	if !found || err != nil {
		return nil
	}
	return val
}

func Test_parsePipelineRunTpl(t *testing.T) {
	type args struct {
		data map[string]interface{}
	}
	tests := []struct {
		name              string
		args              args
		pipelineRunAssert func(obj *unstructured.Unstructured)
		wantErr           bool
	}{{
		name: "Without parameters",
		args: args{
			data: map[string]interface{}{
				"name":      "fake_name",
				"namespace": "fake_namespace",
			},
		},
		pipelineRunAssert: func(obj *unstructured.Unstructured) {
			assert.Equal(t, "fake_name", obj.GetGenerateName())
			assert.Equal(t, "fake_namespace", obj.GetNamespace())
			assert.Equal(t, "fake_name", getNestedString(obj.Object, "spec", "pipelineRef", "name"))
			assert.Equal(t, 0, len(getNestSlice(obj.Object, "spec", "parameters")))
		},
	}, {
		name: "With nil parameters",
		args: args{
			data: map[string]interface{}{
				"name":       "fake_name",
				"namespace":  "fake_namespace",
				"parameters": nil,
			},
		},
		pipelineRunAssert: func(obj *unstructured.Unstructured) {
			assert.Equal(t, "fake_name", obj.GetGenerateName())
			assert.Equal(t, "fake_namespace", obj.GetNamespace())
			assert.Equal(t, "fake_name", getNestedString(obj.Object, "spec", "pipelineRef", "name"))
			assert.Equal(t, 0, len(getNestSlice(obj.Object, "spec", "parameters")))
		},
	}, {
		name: "With empty parameters",
		args: args{
			data: map[string]interface{}{
				"name":       "fake_name",
				"namespace":  "fake_namespace",
				"parameters": map[string]string{},
			},
		},
		pipelineRunAssert: func(obj *unstructured.Unstructured) {
			assert.Equal(t, "fake_name", obj.GetGenerateName())
			assert.Equal(t, "fake_namespace", obj.GetNamespace())
			assert.Equal(t, "fake_name", getNestedString(obj.Object, "spec", "pipelineRef", "name"))
			assert.Equal(t, 0, len(getNestSlice(obj.Object, "spec", "parameters")))
		},
	}, {
		name: "With one parameter",
		args: args{
			data: map[string]interface{}{
				"name":      "fake_name",
				"namespace": "fake_namespace",
				"parameters": map[string]string{
					"a": "b",
				},
			},
		},
		pipelineRunAssert: func(obj *unstructured.Unstructured) {
			assert.Equal(t, "fake_name", obj.GetGenerateName())
			assert.Equal(t, "fake_namespace", obj.GetNamespace())
			assert.Equal(t, "fake_name", getNestedString(obj.Object, "spec", "pipelineRef", "name"))
			assert.Equal(t, 1, len(getNestSlice(obj.Object, "spec", "parameters")))
			assert.Equal(t, map[string]interface{}{"name": "a", "value": "b"}, getNestSlice(obj.Object, "spec", "parameters")[0])
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPipelineRunYaml, err := parsePipelineRunTpl(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePipelineRunTpl() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			obj := unstructured.Unstructured{}
			err = yaml.Unmarshal([]byte(gotPipelineRunYaml), &obj)
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePipelineRunTpl() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			tt.pipelineRunAssert(&obj)
		})
	}
}
