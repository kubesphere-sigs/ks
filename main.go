package main

import (
	"context"
	"fmt"
	ext "github.com/linuxsuren/cobra-extension"
	alias "github.com/linuxsuren/go-cli-alias/pkg"
	aliasCmd "github.com/linuxsuren/go-cli-alias/pkg/cmd"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func main() {
	cmd := &cobra.Command{
		Use: "ks",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			env := os.Environ()

			var gitBinary string
			if gitBinary, err = exec.LookPath("kubectl"); err == nil {
				syscall.Exec(gitBinary, append([]string{"kubectl"}, args...), env)
			}
			return
		},
	}

	var ctx context.Context
	if defMgr, err := alias.GetDefaultAliasMgrWithNameAndInitialData(cmd.Name(), getDefault()); err == nil {
		ctx = context.WithValue(context.Background(), alias.AliasKey, defMgr)

		cmd.AddCommand(aliasCmd.NewRootCommand(ctx))
		aliasCmd.RegisterAliasCompletion(ctx, cmd)
	} else {
		cmd.Println(fmt.Errorf("cannot get default alias manager, error: %v", err))
	}

	cmd.AddCommand(ext.NewCompletionCmd(cmd))

	cmd.SilenceErrors = true
	err := cmd.Execute()
	if err != nil && strings.Contains(err.Error(), "unknown command") {
		args := os.Args[1:]
		var ctx context.Context
		var defMgr *alias.DefaultAliasManager
		var err error
		if defMgr, err = alias.GetDefaultAliasMgrWithNameAndInitialData(cmd.Name(), getDefault()); err == nil {
			ctx = context.WithValue(context.Background(), alias.AliasKey, defMgr)
			if ok, redirect := aliasCmd.RedirectToAlias(ctx, args); ok {
				env := os.Environ()
				var gitBinary string
				if gitBinary, err = exec.LookPath("kubectl"); err == nil {
					syscall.Exec(gitBinary, append([]string{"kubectl"}, redirect...), env)
				}
			} else {
				env := os.Environ()
				var gitBinary string
				if gitBinary, err = exec.LookPath("kubectl"); err == nil {
					syscall.Exec(gitBinary, append([]string{"kubectl"}, args...), env)
				}
			}
		} else {
			err = fmt.Errorf("cannot get default alias manager, error: %v", err)
		}
	}
}

func getDefault() []alias.Alias {
	return []alias.Alias{{
		Name: "pod", Command: "-n kubesphere-system get pod -w",
	}, {
		Name: "j-on", Command: "-n kubesphere-devops-system scale deploy/ks-jenkins --replicas=1",
	}, {
		Name: "j-off", Command: "-n kubesphere-devops-system scale deploy/ks-jenkins --replicas=0",
	}, {
		Name: "j-log", Command: "-n kubesphere-devops-system logs deploy/ks-jenkins --tail=50 -f",
	}, {
		Name: "ctl-log", Command: "-n kubesphere-system logs deploy/ks-controller-manager --tail 50 -f",
	}, {
		Name: "ctl-reset", Command: `-n kubesphere-system patch deploy ks-controller-manager --type=json -p='[{"op":"replace","path":"/spec/template/spec/containers/0/image","value":"kubesphere/ks-controller-manager:v3.0.0"}]'`,
	}, {
		Name: "ctl-reset-dev", Command: `-n kubesphere-system patch deploy ks-controller-manager --type=json -p='[{"op":"replace","path":"/spec/template/spec/containers/0/image","value":"kubespheredev/ks-controller-manager:latest"}]'`,
	}, {
		Name: "api-log", Command: "-n kubesphere-system logs deploy/ks-apiserver --tail 50 -f",
	}, {
		Name: "api-reset", Command: `-n kubesphere-system patch deploy ks-apiserver --type=json -p='[{"op":"replace","path":"/spec/template/spec/containers/0/image","value":"kubesphere/ks-apiserver:v3.0.0"}]'`,
	}, {
		Name: "api-reset-dev", Command: `-n kubesphere-system patch deploy ks-apiserver --type=json -p='[{"op":"replace","path":"/spec/template/spec/containers/0/image","value":"kubespheredev/ks-apiserver:latest"}]'`,
	}, {
		Name: "devops-enable", Command: `-n kubesphere-system patch cc ks-installer -p '{"spec":{"devops":{"enabled":true}}}' --type="merge"`,
	}, {
		Name: "devops-disable", Command: `-n kubesphere-system patch cc ks-installer -p '{"spec":{"devops":{"enabled":false}}}' --type="merge"`,
	}, {
		Name: "install-log", Command: `-n kubesphere-system logs deploy/ks-installer --tail 50 -f`,
	}, {
		Name: "console-reset", Command: `-n kubesphere-system patch deploy ks-console --type=json -p='[{"op":"replace","path":"/spec/template/spec/containers/0/image","value":"kubesphere/ks-console:v3.0.0"}]'`,
	}, {
		Name: "console-reset-dev", Command: `-n kubesphere-system patch deploy ks-console --type=json -p='[{"op":"replace","path":"/spec/template/spec/containers/0/image","value":"kubespheredev/ks-console:latest"}]'`,
	}}
}
