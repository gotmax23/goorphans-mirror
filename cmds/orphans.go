package cmds

import (
	"context"
	"fmt"
	"maps"
	"path"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"go.gtmx.me/goorphans/actions"
	"go.gtmx.me/goorphans/common"
	"go.gtmx.me/goorphans/fasjson"
)

const DefaultURL = "https://a.gtmx.me/orphans/"

var orphansArgsKey = &argsKeyType{"orphans"}

type OrphansArgs struct {
	Dir string
	// TTL float64
}

func newOrphansCommand() *cobra.Command {
	args := &OrphansArgs{}
	cmd := &cobra.Command{
		Use:     "orphans",
		Aliases: []string{"o"},
		Short:   "Subcommands relating to Orphaned Packages Process",
		PersistentPreRun: func(cmd *cobra.Command, argv []string) {
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
	// cmd.PersistentFlags().
	// 	Float64Var(&args.TTL, "ttl", fasjson.DefaultTTL, "Cache TTL in seconds")
	cmd.AddCommand(oDownload())
	cmd.AddCommand(oAddrs())
	return cmd
}

func oDownload() *cobra.Command {
	var baseurl string
	cmd := &cobra.Command{
		Use:     "download",
		Aliases: []string{"d"},
		Short:   "Download orphans data from URL to --dir",
		RunE: func(cmd *cobra.Command, argv []string) error {
			rargs := cmd.Context().Value(rootArgsKey).(*RootArgs)
			args := cmd.Context().Value(orphansArgsKey).(*OrphansArgs)
			return actions.Download(rargs.HTTPClient, baseurl, args.Dir)
		},
		Args: NoArgs,
	}
	cmd.Flags().StringVar(&baseurl, "url", DefaultURL, "Baseurl")
	return cmd
}

func loadOrphans(dir string) (*common.Orphans, error) {
	return common.LoadOrphans(path.Join(dir, common.OrphansJSON))
}

func emails(cache *fasjson.EmailCacheClient, data *common.Orphans) ([]string, error) {
	emailm, err := cache.GetIterEmailsMap(maps.Keys(data.AllAffectedPeople))
	if err != nil {
		return []string{}, err
	}
	return slices.Sorted(maps.Values(emailm)), nil
}

func oAddrs() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "addrs",
		Short: "Get email address for all_affected_people",
		RunE: func(cmd *cobra.Command, argv []string) error {
			rargs := cmd.Context().Value(rootArgsKey).(*RootArgs)
			args := cmd.Context().Value(orphansArgsKey).(*OrphansArgs)
			c, err := rargs.FASCache()
			if err != nil {
				return err
			}
			data, err := loadOrphans(args.Dir)
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
