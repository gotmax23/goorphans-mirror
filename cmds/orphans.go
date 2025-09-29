package cmds

import (
	"context"
	"fmt"
	"maps"
	"os"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"go.gtmx.me/goorphans/actions"
	"go.gtmx.me/goorphans/common"
	"go.gtmx.me/goorphans/config"
	"go.gtmx.me/goorphans/fasjson"
)

var orphansArgsKey = &argsKeyType{"orphans"}

type OrphansArgs struct {
	Dir         string
	RootArgs    *RootArgs
	orphansData *common.Orphans
	Config      *config.OrphansConfig
}

func (args *OrphansArgs) OrphansData() (*common.Orphans, error) {
	if args.orphansData != nil {
		return args.orphansData, nil
	}
	if args.Config.Download {
		o, err := actions.DownloadWithOrphans(
			args.RootArgs.HTTPClient,
			args.Config.BaseURL,
			args.Dir,
		)
		if err != nil {
			return o, err
		}
		args.orphansData = o
		return o, nil
	}
	return common.LoadOrphans(path.Join(args.Dir, common.OrphansJSON))
}

func newOrphansCommand() *cobra.Command {
	var baseurl string
	var download bool
	args := &OrphansArgs{}
	cmd := &cobra.Command{
		Use:     "orphans",
		Aliases: []string{"o"},
		Short:   "Subcommands relating to Orphaned Packages Process",
		PersistentPreRun: func(cmd *cobra.Command, argv []string) {
			rargs := cmd.Context().Value(rootArgsKey).(*RootArgs)
			args.Config = &rargs.Config.Orphans
			if cmd.Flags().Changed("baseurl") {
				args.Config.BaseURL = baseurl
			}
			if cmd.Flags().Changed("download") {
				args.Config.Download = download
			}
			args.RootArgs = rargs
			cmd.SetContext(context.WithValue(cmd.Context(), orphansArgsKey, args))
		},
	}
	pflags := cmd.PersistentFlags()
	pflags.StringVarP(
		&args.Dir,
		"dir",
		"d",
		"orphans",
		"Directory containing orphans.txt and orphans.json",
	)
	pflags.StringVar(&baseurl, "baseurl", "", "Baseurl")
	pflags.BoolVarP(&download, "download", "r", false, "Download new orphans data")
	cmd.AddCommand(oDownload())
	cmd.AddCommand(oAddrs())
	cmd.AddCommand(oLastUpdated())
	cmd.AddCommand(oList())
	return cmd
}

func oDownload() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "download",
		Aliases: []string{"d"},
		Short:   "Download orphans data from baseurl to --dir",
		RunE: func(cmd *cobra.Command, argv []string) error {
			args := cmd.Context().Value(orphansArgsKey).(*OrphansArgs)
			args.Config.Download = true
			d, err := args.OrphansData()
			if err != nil {
				return err
			}
			lastUpdated(d, os.Stderr)
			return nil
		},
		Args: NoArgs,
	}
	return cmd
}

func lastUpdated(d *common.Orphans, file *os.File) *time.Duration {
	if d.FinishedAt != nil {
		elapsed := time.Since(*d.FinishedAt)
		// This data is supposed to be updated once an hour, so just print minutes.
		_, err := fmt.Fprintf(file, "Data was refreshed %.0f minutes ago\n", elapsed.Minutes())
		if err != nil {
			panic(err)
		}
		return &elapsed
	}
	return nil
}

func emails(cache *fasjson.EmailCacheClient, data *common.Orphans) ([]string, error) {
	emailm, err := cache.GetIterEmailsMap(maps.Keys(data.AllAffectedPeople))
	if err != nil {
		return []string{}, err
	}
	return slices.Sorted(maps.Values(emailm)), nil
}

func oLastUpdated() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "last-updated",
		Short: "Load the local orphans data and how long it's been since the last update",
		RunE: func(cmd *cobra.Command, argv []string) error {
			args := cmd.Context().Value(orphansArgsKey).(*OrphansArgs)
			d, err := args.OrphansData()
			if err != nil {
				return err
			}
			duration := lastUpdated(d, os.Stdout)
			if duration == nil {
				return fmt.Errorf("finished_at was not included in the orphans data")
			}
			return nil
		},
	}
	return cmd
}

func oAddrs() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "addrs",
		Short: "Get email address for all_affected_people",
		RunE: func(cmd *cobra.Command, argv []string) error {
			args := cmd.Context().Value(orphansArgsKey).(*OrphansArgs)
			c, err := args.RootArgs.FASCache()
			if err != nil {
				return err
			}
			data, err := args.OrphansData()
			if err != nil {
				return err
			}
			emails, err := emails(c, data)
			if err != nil {
				return err
			}
			fmt.Println(strings.Join(emails, "\n"))
			return nil
		},
		Args: NoArgs,
	}
	return cmd
}

var completeGolangExemption = cobra.FixedCompletions(
	slices.Collect(maps.Keys(common.ToGolangExemption)),
	cobra.ShellCompDirectiveNoFileComp,
)

func oList() *cobra.Command {
	var out string
	ge := common.GolangExemptionMust
	weeks := 6
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List orphans",
		RunE: func(cmd *cobra.Command, argv []string) error {
			args := cmd.Context().Value(orphansArgsKey).(*OrphansArgs)
			o, err := args.OrphansData()
			if err != nil {
				return err
			}
			r, err := o.OrphanedFilter(
				common.OrphanedFilterOptions{
					Duration:        common.Weeks(weeks),
					GolangExemption: ge,
				},
			)
			if err != nil {
				return err
			}
			return common.WriteFileLines(out, r)
		},
	}
	cmd.Flags().
		StringVarP(&out, "output", "o", "-", "Output file; defaults to stdout")
	cmd.Flags().IntVarP(&weeks, "weeks", "w", weeks, "")
	cmd.Flags().TextVar(&ge, "golang-exemption", ge, "must (default), optional, or ignore")
	_ = cmd.RegisterFlagCompletionFunc("golang-exemption", completeGolangExemption)
	return cmd
}

// func oSomething() *cobra.Command {
// 	cmd := &cobra.Command{
// 		Use:   "something",
// 		Short: "Do something",
// 		RunE: func(cmd *cobra.Command, argv []string) error {
// 			args := cmd.Context().Value(orphansArgsKey).(*OrphansArgs)
// 			return nil
// 		},
// 	}
// 	return cmd
// }
