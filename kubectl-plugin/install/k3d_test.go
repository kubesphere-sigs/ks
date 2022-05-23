package install

import (
	"fmt"
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

func Test_newInstallK3DCmd(t *testing.T) {
	cmd := newInstallK3DCmd()
	assert.NotNil(t, cmd)
	assert.Equal(t, "k3d", cmd.Use)

}

func Test_isBiggerThan(t *testing.T) {
	type args struct {
		current string
		target  string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr assert.ErrorAssertionFunc
	}{{
		name: "v1.24.0 > v1.23.0",
		args: args{
			current: "v1.24.0",
			target:  "v1.23.0",
		},
		want: true,
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			return false
		},
	}, {
		name: "v1.24.0 < v1.23.0",
		args: args{
			current: "v1.23.0",
			target:  "v1.24.0",
		},
		want: false,
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			return false
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := isBiggerThan(tt.args.current, tt.args.target)
			if !tt.wantErr(t, err, fmt.Sprintf("isBiggerThan(%v, %v)", tt.args.current, tt.args.target)) {
				return
			}
			assert.Equalf(t, tt.want, got, "isBiggerThan(%v, %v)", tt.args.current, tt.args.target)
		})
	}
}
