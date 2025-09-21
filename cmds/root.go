package cmds

import (
	"context"
	"net/http"
	"os"
	"path"

	"github.com/spf13/cobra"
	"go.gtmx.me/goorphans/common"
	"go.gtmx.me/goorphans/fasjson"
)

type argsKeyType struct{ name string }

var rootArgsKey = &argsKeyType{"root"}

type RootArgs struct {
	HTTPClient *http.Client
	CacheDir   string
	TTL        float64
	DBPath     string
	fasCache   *fasjson.EmailCacheClient
}

func (args *RootArgs) FASCache() (*fasjson.EmailCacheClient, error) {
	if args.fasCache != nil {
		return args.fasCache, nil
	}
	if args.DBPath == "" {
		args.DBPath = path.Join(args.CacheDir, "fasjson.db")
	}
	c, err := fasjson.OpenCacheDB(args.DBPath, args.TTL)
	if err != nil {
		return nil, err
	}
	return c, nil
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
		PersistentPreRunE: func(cmd *cobra.Command, argv []string) error {
			args.HTTPClient = &http.Client{}
			c, err := common.CacheDir()
			if err != nil {
				return err
			}
			args.CacheDir = c
			cmd.SetContext(context.WithValue(cmd.Context(), rootArgsKey, &args))
			return nil
		},
		SilenceUsage: true,
	}
	rootCmd.PersistentFlags().
		Float64Var(&args.TTL, "fasjson-ttl", fasjson.DefaultTTL, "TTL for the FASJSON cache")
	rootCmd.PersistentFlags().
		StringVar(
			&args.DBPath, "fasjson-db", "",
			"Path to cache database. Defaults to $XDG_CACHE_HOME/goorphans/fasjson.db",
		)
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
