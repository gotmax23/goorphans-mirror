package cmds

import (
	"context"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

type argsKeyType struct{ name string }

var rootArgsKey = &argsKeyType{"root"}

type RootArgs struct {
	HTTPClient *http.Client
}

// ArgsWrapper wraps Cobra.PositionalArgs functions and prints usage
// information if there is an error.
// This way, we can set SilenceUsage for other errors but still print the usage
// for positional args errors.
func ArgsWrapper(f cobra.PositionalArgs) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		err := f(cmd, args)
		if err != nil {
			cmd.PrintErrln(cmd.UsageString())
		}
		return err
	}
}

// NoArgs wraps [cobra.NoArgs] with [ArgsWrapper]
var NoArgs = ArgsWrapper(cobra.NoArgs)

func NewRootCmd() *cobra.Command {
	cobra.EnableTraverseRunHooks = true
	args := RootArgs{}
	rootCmd := &cobra.Command{
		Use:   "goorphans",
		Short: "Manage the Fedora orphaned packges process announcements",
		PersistentPreRun: func(cmd *cobra.Command, argv []string) {
			// rclient := retryablehttp.NewClient()
			// args.HTTPClient = rclient.StandardClient()
			args.HTTPClient = &http.Client{}
			cmd.SetContext(context.WithValue(cmd.Context(), rootArgsKey, &args))
		},
		SilenceUsage: true,
	}
	rootCmd.AddCommand(newOrphansCommand())
	rootCmd.AddCommand(newFas2emailCommand())
	rootCmd.AddCommand(NewDistgitCmd())
	return rootCmd
}

// rootCmd represents the base command when called without any subcommands

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := NewRootCmd().Execute()
	if err != nil {
		os.Exit(1)
	}
}
