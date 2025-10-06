package cmds

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/fatih/color"
	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"
	"go.gtmx.me/goorphans/config"
	"go.gtmx.me/goorphans/fasjson"
)

type argsKeyType struct{ name string }

var rootArgsKey = &argsKeyType{"root"}

type RootArgs struct {
	HTTPClient *http.Client
	Config     *config.Config
	fasCache   *fasjson.EmailCacheClient
}

func (args *RootArgs) FASCache() (*fasjson.EmailCacheClient, error) {
	if args.fasCache != nil {
		return args.fasCache, nil
	}
	c, err := fasjson.OpenCacheDB(args.Config.FASJSON.DB, args.Config.FASJSON.TTL)
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
	var configPath string
	var ttl float64
	var dbPath string
	cobra.EnableTraverseRunHooks = true
	args := RootArgs{}
	rootCmd := &cobra.Command{
		Use:   "goorphans",
		Short: "Manage the Fedora orphaned packges process announcements",
		PersistentPreRunE: func(cmd *cobra.Command, argv []string) error {
			if cmd.Name() == "__complete" {
				return nil
			}
			// TODO: Either document that --config="" disables config loading
			// or remove this behaivor.
			if !cmd.Root().PersistentFlags().Changed("config") {
				configPath = config.DefaultSentinel
			}
			env, ok := os.LookupEnv("GOORPHANS_CONFIG")
			if ok {
				configPath = env
			}
			config, err := config.LoadConfig(configPath)
			if err != nil {
				return err
			}
			args.Config = config
			if cmd.Root().PersistentFlags().Changed("fasjson-ttl") {
				args.Config.FASJSON.TTL = ttl
			}
			if cmd.Root().PersistentFlags().Changed("fasjson-db") {
				args.Config.FASJSON.DB = dbPath
			}
			// err = args.Config.SMTP.Validate()
			// if err != nil {
			// 	return err
			// }
			args.HTTPClient = &http.Client{}
			cmd.SetContext(context.WithValue(cmd.Context(), rootArgsKey, &args))
			return nil
		},
		SilenceUsage: true,
	}
	rootCmd.PersistentFlags().
		StringVarP(
			&configPath, "config", "c", "",
			"Path to config. Defaults to $XDG_CONFIG_HOME/goorphans.toml.",
		)
	rootCmd.PersistentFlags().
		Float64Var(&ttl, "fasjson-ttl", 0, "TTL for the FASJSON cache")
	rootCmd.PersistentFlags().
		StringVar(
			&dbPath, "fasjson-db", "",
			"Path to cache database. Defaults to $XDG_CACHE_HOME/goorphans/fasjson.db",
		)
	rootCmd.AddCommand(newOrphansCommand())
	rootCmd.AddCommand(newFas2emailCommand())
	rootCmd.AddCommand(NewDistgitCmd())
	rootCmd.AddCommand(newDumpConfigCmd())
	// rootCmd.AddCommand(newDocsGenCmd())
	return rootCmd
}

func newDumpConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dump-config",
		Short: "Dump configuration as TOML",
		RunE: func(cmd *cobra.Command, argv []string) error {
			rargs := cmd.Context().Value(rootArgsKey).(*RootArgs)
			if err := rargs.Config.SMTP.Validate(); err != nil {
				colorToStderrForce(color.FgRed, "%v\n", err)
			}
			if rargs.Config.SMTP.Password != "" {
				rargs.Config.SMTP.Password = "REDACTED"
			}
			b, err := toml.Marshal(rargs.Config)
			if err != nil {
				return err
			}
			fmt.Println(string(b))
			return nil
		},
	}
	return cmd
}

// TODO: Playing around with docs gen
// func newDocsGenCmd() *cobra.Command {
// 	out := "docs"
// 	cmd := &cobra.Command{
// 		Use:    "_docs",
// 		Hidden: true,
// 		RunE: func(cmd *cobra.Command, argv []string) error {
// 			if err := os.Mkdir(out, 0o755); err != nil && !errors.Is(err, os.ErrExist) {
// 				return err
// 			}
// 			root := NewRootCmd()
// 			root.DisableAutoGenTag = true
// 			return doc.GenMarkdownTree(root, out)
// 		},
// 	}
// 	cmd.Flags().StringVarP(&out, "out", "o", out, "")
// 	return cmd
// }

func Execute() {
	err := NewRootCmd().Execute()
	if err != nil {
		os.Exit(1)
	}
}
