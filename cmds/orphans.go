package cmds

import (
	"context"
	"fmt"
	"maps"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"go.gtmx.me/goorphans/actions"
	"go.gtmx.me/goorphans/common"
	"go.gtmx.me/goorphans/fasjson"
)

const DefaultURL = "https://a.gtmx.me/orphans/"

var orphansArgsKey = &argsKeyType{"orphans"}

type OrphansArgs struct {
	BaseURL     string
	Dir         string
	Download    bool
	RootArgs    *RootArgs
	orphansData *common.Orphans
}

func (args *OrphansArgs) OrphansData() (*common.Orphans, error) {
	if args.orphansData != nil {
		return args.orphansData, nil
	}
	if args.Download {
		o, err := actions.DownloadWithOrphans(args.RootArgs.HTTPClient, args.BaseURL, args.Dir)
		if err != nil {
			return o, err
		}
		args.orphansData = o
		return o, nil
	}
	return common.LoadOrphans(path.Join(args.Dir, common.OrphansJSON))
}

func newOrphansCommand() *cobra.Command {
	args := &OrphansArgs{}
	cmd := &cobra.Command{
		Use:     "orphans",
		Aliases: []string{"o"},
		Short:   "Subcommands relating to Orphaned Packages Process",
		PersistentPreRun: func(cmd *cobra.Command, argv []string) {
			rargs := cmd.Context().Value(rootArgsKey).(*RootArgs)
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
	pflags.StringVar(&args.BaseURL, "baseurl", DefaultURL, "Baseurl")
	pflags.BoolVarP(&args.Download, "download", "r", false, "Download new orphans data")
	cmd.AddCommand(oDownload())
	cmd.AddCommand(oAddrs())
	cmd.AddCommand(oLastUpdated())
	return cmd
}

func oDownload() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "download",
		Aliases: []string{"d"},
		Short:   "Download orphans data from baseurl to --dir",
		RunE: func(cmd *cobra.Command, argv []string) error {
			args := cmd.Context().Value(orphansArgsKey).(*OrphansArgs)
			args.Download = true
			d, err := args.OrphansData()
			if err != nil {
				return err
			}
			_, _ = lastUpdated(d, false)
			return nil
		},
		Args: NoArgs,
	}
	return cmd
}

func lastUpdated(d *common.Orphans, warn bool) (*time.Duration, error) {
	if d.FinishedAt != nil {
		elapsed := time.Since(*d.FinishedAt)
		// This data is supposed to be updated once an hour, so just print minutes.
		fmt.Printf("Data was refreshed %.0f minutes ago\n", elapsed.Minutes())
		return &elapsed, nil
	}
	if warn {
		return nil, fmt.Errorf("finished_at was not included in the orphans data")
	}
	return nil, nil
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
			_, err = lastUpdated(d, true)
			return err
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
