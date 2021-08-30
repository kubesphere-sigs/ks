package entrypoint

import (
	"fmt"
	"github.com/spf13/cobra"
)

func GetCmdPath(cmd *cobra.Command) string {
	current := cmd.Use
	if cmd.HasParent() {
		parentName := GetCmdPath(cmd.Parent())
		if parentName == "" {
			return current
		}

		return fmt.Sprintf("%s.%s", parentName, current)
	}
	// don't need the name of root cmd
	return ""
}
