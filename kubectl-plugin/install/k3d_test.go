package install

import (
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
