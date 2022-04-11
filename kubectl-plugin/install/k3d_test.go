package install

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_isGreaterThanV5(t *testing.T) {
	type args struct {
		version string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "k3d v4",
			args: args{
				"v4.4.8",
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "k3d v5",
			args: args{
				"v5.0.0",
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := isGreaterThanV5(tt.args.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkK3dVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("checkK3dVersion() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getRegistryConfig(t *testing.T) {
	type args struct {
		regMap map[string]string
	}
	tests := []struct {
		name string
		args args
		want string
	}{{
		name: "normal case",
		args: args{regMap: map[string]string{
			"registry":      "docker.io",
			"registry-ghcr": "ghcr.io",
		}},
		want: `mirrors:
  docker.io:
    endpoint:
    - http://k3d-registry:5000
  k3d-registry:5000:
    endpoint:
    - http://k3d-registry:5000
  ghcr.io:
    endpoint:
    - http://k3d-registry-ghcr:5000
  k3d-registry-ghcr:5000:
    endpoint:
    - http://k3d-registry-ghcr:5000
configs: {}
auths: {}`,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, getRegistryConfig(tt.args.regMap), "getRegistryConfig(%v)", tt.args.regMap)
		})
	}
}
