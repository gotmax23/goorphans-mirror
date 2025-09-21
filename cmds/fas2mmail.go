package cmds

import (
	"context"
	"fmt"
	"maps"
	"path"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"go.gtmx.me/goorphans/common"
	"go.gtmx.me/goorphans/fasjson"
)

var fas2emailArgsKey = &argsKeyType{"fas2email"}

type Fas2emailArgs struct {
	Cache *fasjson.EmailCacheClient
}

func Fas2emailCache() (string, error) {
	cacheDir, err := common.CacheDir()
	if err != nil {
		return "", err
	}
	return path.Join(cacheDir, "fasjson.db"), err
}

func newFas2emailCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "fas2email",
		Short:   "Query FASJSON for user and group emails",
		Aliases: []string{"f2e"},
		PersistentPreRunE: func(cmd *cobra.Command, argv []string) error {
			rargs := cmd.Context().Value(rootArgsKey).(*RootArgs)
			cache, err := rargs.FASCache()
			if err != nil {
				return err
			}
			args := &Fas2emailArgs{cache}
			cmd.SetContext(context.WithValue(cmd.Context(), fas2emailArgsKey, args))
			return nil
		},
	}
	cmd.AddCommand(f2eClean())
	cmd.AddCommand(f2eGet())
	cmd.AddCommand(f2eGetFile())
	return cmd
}

func f2eClean() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Init cache database and clean old cache",
		RunE: func(cmd *cobra.Command, argv []string) error {
			args := cmd.Context().Value(fas2emailArgsKey).(*Fas2emailArgs)
			fmt.Println("Cleaning cache...")
			return args.Cache.Clean()
		},
		Args: NoArgs,
	}
	return cmd
}

func f2eGet() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get NAME...",
		Short: "Get emails for usernames or group names (args prefixed with @)",
		RunE: func(cmd *cobra.Command, argv []string) error {
			args := cmd.Context().Value(fas2emailArgsKey).(*Fas2emailArgs)
			emailm, err := args.Cache.GetAllEmailsMap(argv)
			if err != nil {
				return err
			}
			emails := slices.Sorted(maps.Values(emailm))
			fmt.Println(strings.Join(emails, "\n"))
			return nil
		},
		Args: ArgsWrapper(cobra.MinimumNArgs(1)),
	}
	return cmd
}

func f2eGetFile() *cobra.Command {
	var in, out string
	cmd := &cobra.Command{
		Use:     "get-file",
		Aliases: []string{"getf"},
		Short:   "Get emails for usernames or group names (name prefixed with @)",
		RunE: func(cmd *cobra.Command, argv []string) error {
			args := cmd.Context().Value(fas2emailArgsKey).(*Fas2emailArgs)
			names, err := common.ReadFileLines(in)
			if err != nil {
				return err
			}
			emailm, err := args.Cache.GetAllEmailsMap(names)
			if err != nil {
				return err
			}
			emails := slices.Sorted(maps.Values(emailm))
			return common.WriteFileLines(out, emails)
		},
		Args: NoArgs,
	}
	cmd.Flags().
		StringVarP(
			&in, "input", "i", "-",
			"Input file containing newline-separated list; defaults to stdin",
		)
	cmd.Flags().
		StringVarP(&out, "output", "o", "-", "Output file; defaults to stdout")
	return cmd
}
