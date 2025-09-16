package cmds

import (
	"github.com/spf13/cobra"
	"go.gtmx.me/goorphans/actions"
)

const DefaultURL = "https://a.gtmx.me/orphans/"

func newDownloadCmd() *cobra.Command {
	var baseurl string
	cmd := &cobra.Command{
		Use:          "download",
		Short:        "Download orphans data from URL",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, argv []string) error {
			rargs := cmd.Context().Value(rootArgsKey).(*RootArgs)
			return actions.Download(rargs.HTTPClient, baseurl, rargs.Dir)
		},
		Args: NoArgs,
	}
	cmd.Flags().StringVar(&baseurl, "url", DefaultURL, "Baseurl")
	return cmd
}
