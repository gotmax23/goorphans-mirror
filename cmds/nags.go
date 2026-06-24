package cmds

import (
	"github.com/spf13/cobra"
	"go.gtmx.me/goorphans/actions"
)

func newNagsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nags",
		Short: "Send reminder emails to Fedora packagers for various purposes",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			rargs := cmd.Context().Value(rootArgsKey).(*RootArgs)
			if err := rargs.Config.SMTP.Validate(); err != nil {
				return err
			}
			return nil
		},
	}
	cmd.AddCommand(nags2FA())
	return cmd
}

func nags2FA() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "2fa PATH",
		Args: ArgsWrapper(cobra.ExactArgs(1)),
		RunE: func(cmd *cobra.Command, argv []string) error {
			rargs := cmd.Context().Value(rootArgsKey).(*RootArgs)
			return actions.Send2FANag(cmd.Context(), rargs.Config, argv[0])
		},
	}
	return cmd
}
