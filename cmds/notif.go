package cmds

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newNotifCmd() *cobra.Command {
	notifCmd := &cobra.Command{
		Use:   "notif",
		Short: "Send individual notifications",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("TODO")
		},
	}
	return notifCmd
}
