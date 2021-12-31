package config

import (
	"reflect"
	"testing"
)

func Test_patchKubeSphereConfig(t *testing.T) {
	type args struct {
		kubeSphereConfig map[string]interface{}
		patch            map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]interface{}
		wantErr bool
	}{{
		name: "Patch non-exist fields recursively",
		args: args{
			kubeSphereConfig: map[string]interface{}{
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
		name: "Patch non-exist field",
		args: args{
			kubeSphereConfig: map[string]interface{}{
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
		name: "Patch existing field",
		args: args{
			kubeSphereConfig: map[string]interface{}{
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
		name: "Report an error if patch is nil",
		args: args{
			kubeSphereConfig: map[string]interface{}{
				"devops": map[string]interface{}{
					"enabled": true,
				},
			},
			patch: nil,
		},
		wantErr: true,
	}, {
		name: "Report an error if KubeSphereConfig is nil",
		args: args{
			kubeSphereConfig: nil,
			patch:            map[string]interface{}{},
		},
		wantErr: true,
	}, {
		name: "Patch with different type",
		args: args{
			kubeSphereConfig: map[string]interface{}{
				"devops": map[string]interface{}{
					"enabled": true,
				},
			},
			patch: map[string]interface{}{
				"devops": "awesome",
			},
		},
		want: map[string]interface{}{
			"devops": "awesome",
		},
	}, {
		name: "Patch with map[string]string type",
		args: args{
			kubeSphereConfig: map[string]interface{}{
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
				"password": "patch password",
			},
		},
	}, {
		name: "Patch without map[string]interface{}",
		args: args{
			kubeSphereConfig: map[string]interface{}{
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
			"devops": map[string]interface{}{
				"enabled":  true,
				"disabled": false,
			},
		},
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := patchKubeSphereConfig(tt.args.kubeSphereConfig, tt.args.patch)
			if (err != nil) != tt.wantErr {
				t.Errorf("patchKubeSphereConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("patchKubeSphereConfig() = %+v, want =  %+v", got, tt.want)
			}
		})
	}
}
