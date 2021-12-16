package entrypoint

import (
	"github.com/spf13/cobra"
	"testing"
)

func TestGetCmdPath(t *testing.T) {
	type args struct {
		cmd func() *cobra.Command
	}
	tests := []struct {
		name string
		args args
		want string
	}{{
		name: "no sub-command",
		args: args{
			cmd: func() *cobra.Command {
				return &cobra.Command{
					Use: "root",
				}
			},
		},
		want: "",
	}, {
		name: "with one sub-command",
		args: args{
			cmd: func() *cobra.Command {
				root := &cobra.Command{
					Use: "root",
				}
				sub := &cobra.Command{
					Use: "sub1",
				}
				root.AddCommand(sub)
				return sub
			},
		},
		want: "sub1",
	}, {
		name: "with two sub-commands",
		args: args{
			cmd: func() *cobra.Command {
				root := &cobra.Command{
					Use: "root",
				}
				sub1 := &cobra.Command{
					Use: "sub1",
				}
				sub2 := &cobra.Command{
					Use: "sub2",
				}
				root.AddCommand(sub1)
				sub1.AddCommand(sub2)
				return sub2
			},
		},
		want: "sub1.sub2",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetCmdPath(tt.args.cmd()); got != tt.want {
				t.Errorf("GetCmdPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
