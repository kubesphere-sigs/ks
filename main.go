package main

import (
	"context"
	"fmt"
	ext "github.com/linuxsuren/cobra-extension"
	extver "github.com/linuxsuren/cobra-extension/version"
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

	cmd.AddCommand(extver.NewVersionCmd("linuxsuren", "ks", "ks", nil))

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
			var gitBinary string
			var targetCmd []string
			env := os.Environ()

			if gitBinary, err = exec.LookPath("kubectl"); err != nil {
				panic("cannot find kubectl")
			}

			if ok, redirect := aliasCmd.RedirectToAlias(ctx, args); ok {
				targetCmd = append([]string{"kubectl"}, redirect...)

				//fmt.Println("the real command is:", strings.Join(targetCmd, " "))
			} else {
				targetCmd = append([]string{"kubectl"}, args...)
			}
			_ = syscall.Exec(gitBinary, targetCmd, env) // ignore the errors due to we've no power to deal with it
		} else {
			err = fmt.Errorf("cannot get default alias manager, error: %v", err)
		}
	}
}
