package token

import (
	"fmt"
	"github.com/spf13/cobra"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"os/exec"
	"strings"
)

// NewTokenCmd returns the token command
func NewTokenCmd(client dynamic.Interface, clientset *kubernetes.Clientset) (cmd *cobra.Command) {
	opt := &option{}
	cmd = &cobra.Command{
		Use:   "token",
		Short: "Print the kubectl token of KubeSphere",
		RunE:  opt.runE,
	}

	flags := cmd.Flags()
	flags.StringVarP(&opt.Host, "host", "", "",
		"The host address of SSH")
	flags.StringVarP(&opt.User, "user", "u", "root",
		"The user of SSH server")
	flags.IntVarP(&opt.Port, "port", "p", 22,
		"The port of SSH server")
	return
}

// option is the option for token command
type option struct {
	Host string
	User string
	Port int
}

func (o *option) runE(cmd *cobra.Command, args []string) (err error) {
	var name string
	if name, err = o.getSecretName(); err != nil {
		return
	}
	var data []byte
	cmdArgs := fmt.Sprintf("%s@%s -p %d kubectl get %s -n kubesphere-system -ojsonpath={.data.token} | base64 -d", o.User, o.Host, o.Port, name)
	if data, err = exec.Command("ssh", strings.Split(cmdArgs, " ")...).Output(); err == nil {
		cmd.Println(string(data))
	}
	return
}

func (o *option) getSecretName() (name string, err error) {
	var data []byte
	cmd := fmt.Sprintf("%s@%s -p %d kubectl get -n kubesphere-system secret -oname | grep kubesphere-token", o.User, o.Host, o.Port)
	if data, err = exec.Command("ssh", strings.Split(cmd, " ")...).Output(); err == nil {
		name = strings.TrimSpace(string(data))
	}
	return
}
