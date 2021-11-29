package config

import (
	"reflect"
	"testing"
)

func Test_patchKubeSphereConfig(t *testing.T) {
	type args struct {
		main  map[string]interface{}
		patch map[string]interface{}
	}
	tests := []struct {
		name string
		args args
		want map[string]interface{}
	}{{
		name: "Merge non-exist fields recursively",
		args: args{
			main: map[string]interface{}{
				"devops": map[string]interface{}{
					"enabled": true,
				},
			},
			patch: map[string]interface{}{
				"devops": map[string]interface{}{
					"password": "fake password",
				},
			},
		},
		want: map[string]interface{}{
			"devops": map[string]interface{}{
				"enabled":  true,
				"password": "fake password",
			},
		},
	}, {
		name: "Merge non-exist field",
		args: args{
			main: map[string]interface{}{
				"devops": map[string]interface{}{
					"enabled": true,
				},
			},
			patch: map[string]interface{}{
				"sonarQube": map[string]interface{}{
					"token": "fake token",
				},
			},
		},
		want: map[string]interface{}{
			"devops": map[string]interface{}{
				"enabled": true,
			},
			"sonarQube": map[string]interface{}{
				"token": "fake token",
			},
		},
	}, {
		name: "Merge existing field",
		args: args{
			main: map[string]interface{}{
				"devops": map[string]interface{}{
					"enabled": true,
				},
				"sonarQube": map[string]interface{}{
					"token": "fake token",
				},
			},
			patch: map[string]interface{}{
				"devops": map[string]interface{}{
					"enabled": false,
				},
			},
		},
		want: map[string]interface{}{
			"devops": map[string]interface{}{
				"enabled": false,
			},
			"sonarQube": map[string]interface{}{
				"token": "fake token",
			},
		},
	}, {
		name: "Patch is nil",
		args: args{
			main: map[string]interface{}{
				"devops": map[string]interface{}{
					"enabled": true,
				},
			},
			patch: nil,
		},
		want: map[string]interface{}{
			"devops": map[string]interface{}{
				"enabled": true,
			},
		},
	}, {
		name: "Main is nil",
		args: args{
			main:  nil,
			patch: map[string]interface{}{},
		},
		want: nil,
	}, {
		name: "Merge with different type",
		args: args{
			main: map[string]interface{}{
				"devops": map[string]interface{}{
					"enabled": true,
				},
			},
			patch: map[string]interface{}{
				"devops": "awesome",
			},
		},
		want: map[string]interface{}{
			"devops": map[string]interface{}{
				"enabled": true,
			},
		},
	}, {
		name: "Merge with different map type",
		args: args{
			main: map[string]interface{}{
				"devops": map[string]interface{}{
					"enabled":  true,
					"password": "fake password",
				},
			},
			patch: map[string]interface{}{
				"devops": map[string]string{
					"password": "patch password",
				},
			},
		},
		want: map[string]interface{}{
			"devops": map[string]interface{}{
				"enabled":  true,
				"password": "fake password",
			},
		},
	}, {
		name: "Merge without map[string]interface{}",
		args: args{
			main: map[string]interface{}{
				"devops": map[string]bool{
					"enabled": true,
				},
			},
			patch: map[string]interface{}{
				"devops": map[string]bool{
					"disabled": false,
				},
			},
		},
		want: map[string]interface{}{
			"devops": map[string]bool{
				"disabled": false,
			},
		},
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mergeMap(tt.args.main, tt.args.patch)
			if !reflect.DeepEqual(tt.args.main, tt.want) {
				t.Errorf("mergeMap() = %v, want =  %v", tt.args.main, tt.want)
				return
			}
		})
	}
}
