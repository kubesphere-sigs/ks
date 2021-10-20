package install

import (
	"github.com/kubesphere-sigs/ks/kubectl-plugin/types"
	"testing"
)

func Test_checkK3dVersion(t *testing.T) {
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
			name: types.K3dVersion4,
			args: args{
				types.K3dVersion4,
			},
			want:    false,
			wantErr: false,
		},
		{
			name: types.K3dVersion5,
			args: args{
				types.K3dVersion5,
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := checkK3dVersion(tt.args.version)
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
