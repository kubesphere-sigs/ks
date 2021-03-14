package user

import (
	"context"
	"fmt"
	kstype "github.com/linuxsuren/ks/kubectl-plugin/types"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
)

// NewUserCmd returns the command of users
func NewUserCmd(client dynamic.Interface) (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:   "user",
		Short: "Reset the password of Kubesphere to the default value which is same with its name",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			name := args[0]

			_, err = client.Resource(kstype.GetUserSchema()).Patch(context.TODO(),
				name,
				types.MergePatchType,
				[]byte(fmt.Sprintf(`{"spec":{"password":"%s"},"metadata":{"annotations":null}}`, name)),
				metav1.PatchOptions{})
			return
		},
	}
	return
}
