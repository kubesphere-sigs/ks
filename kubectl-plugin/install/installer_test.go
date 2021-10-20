package install

import "testing"

func Test_isNotReleaseVersion(t *testing.T) {
	type args struct {
		version string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{{
		name: "rc version",
		args: args{
			version: "v1.0-rc1",
		},
		want: true,
	}, {
		name: "alpha version",
		args: args{
			version: "v1.0-alpha1",
		},
		want: true,
	}, {
		name: "beta version",
		args: args{
			version: "v1.0-beta-1",
		},
		want: true,
	}, {
		name: "release version",
		args: args{
			version: "v1.0",
		},
		want: false,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isNotReleaseVersion(tt.args.version); got != tt.want {
				t.Errorf("isNotReleaseVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}
