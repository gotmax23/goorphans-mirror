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

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	gomail "github.com/wneessen/go-mail"
	"go.gtmx.me/goorphans/actions"
	"go.gtmx.me/goorphans/common"
	"go.gtmx.me/goorphans/config"
	"go.gtmx.me/goorphans/fasjson"
	"go.gtmx.me/goorphans/mail"
	"go.gtmx.me/goorphans/notifs"
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
	cmd.AddCommand(oAnnounce())
	cmd.AddCommand(oNotifications())
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
		minutes := elapsed.Minutes()
		clr := color.New()
		switch {
		case minutes >= 60:
			clr.Add(color.FgRed)
		case minutes <= 5:
			clr.Add(color.FgGreen)
		default:
			clr.Add(color.FgBlue)
		}
		_, err := clr.Fprintf(file, "Data was refreshed %.0f minutes ago\n", elapsed.Minutes())
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
		Args:  NoArgs,
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
	count := false

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List orphans",
		Args:    NoArgs,
		RunE: func(cmd *cobra.Command, argv []string) error {
			args := cmd.Context().Value(orphansArgsKey).(*OrphansArgs)
			o, err := args.OrphansData()
			if args.Config.Download {
				lastUpdated(o, os.Stderr)
			}
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
			if count {
				fmt.Println(len(r))
			} else {
				err = common.WriteFileLines(out, r)
				if err != nil {
					return err
				}
				colorToStderrF(color.FgMagenta, "    %d orphans listed\n", len(r))
			}
			return nil
		},
	}
	cmd.Flags().
		StringVarP(&out, "output", "o", "-", "Output file; defaults to stdout")
	cmd.Flags().IntVarP(&weeks, "weeks", "w", weeks, "")
	cmd.Flags().TextVar(&ge, "golang-exemption", ge,
		"must (default), optional, ignore, or only",
	)
	cmd.Flags().BoolVar(&count, "count", count, "Only print a count")
	_ = cmd.RegisterFlagCompletionFunc("golang-exemption", completeGolangExemption)
	return cmd
}

// makeAnnounceMsg prepares a [gomail.Msg] struct.
// Set noRecpts to avoid sending to any recipients and only the BCC value from
// the config.
func makeAnnounceMsg(
	config *config.Config,
	o *common.Orphans,
	f *fasjson.EmailCacheClient,
	noRecpts bool,
) (*gomail.Msg, error) {
	affected := o.AllAffectedPeople
	msg := gomail.NewMsg(gomail.WithNoDefaultUserAgent())
	msg.Subject("Orphaned packages looking for new maintainers")

	if config.Orphans.DirectMaintsOnly {
		affected = o.AffectedPeople
	}
	emails, err := f.GetIterEmailsMap(maps.Keys(affected))
	if err != nil {
		return msg, err
	}
	length := len(config.Orphans.BCC)
	if !noRecpts {
		length += len(config.Orphans.BCC)
	}
	bcc := make([]string, 0, length)
	bcc = append(bcc, config.Orphans.BCC...)

	if !noRecpts {
		if err := msg.To(config.Orphans.To...); err != nil {
			return msg, err
		}
		if err := msg.ReplyTo(config.Orphans.ReplyTo); err != nil {
			return msg, err
		}
		for _, email := range emails {
			bcc = append(bcc, email)
		}
	}
	if err := msg.Bcc(bcc...); err != nil {
		return msg, err
	}
	return msg, nil
}

func sendMsg(config *config.Config, msg *gomail.Msg) error {
	err := mail.FinalizeMsg(&config.SMTP, msg)
	if err != nil {
		return err
	}
	r, _ := msg.GetRecipients()
	fmt.Printf("Sending %q to %d recipients...\n", msg.GetGenHeader("Subject")[0], len(r))
	err = msg.WriteToFile("message.eml")
	return err
	// TODO: Actually send messages.
	// c, err := mail.NewClient(&config.SMTP)
	// if err != nil {
	// 	return err
	// }
	// return c.DialAndSend(msg)
}

func oAnnounce() *cobra.Command {
	direct := false
	cmd := &cobra.Command{
		Use:   "announce",
		Short: "Send announcement",
		Args:  NoArgs,
		RunE: func(cmd *cobra.Command, argv []string) error {
			args := cmd.Context().Value(orphansArgsKey).(*OrphansArgs)
			if cmd.Flags().Changed("direct-maints") {
				args.Config.DirectMaintsOnly = direct
			}

			o, err := args.OrphansData()
			if err != nil {
				return err
			}
			lastUpdated(o, os.Stderr)

			f, err := args.RootArgs.FASCache()
			if err != nil {
				return err
			}

			msg, err := makeAnnounceMsg(args.RootArgs.Config, o, f, false)
			if err != nil {
				return err
			}

			p := path.Join(args.Dir, common.OrphansTXT)
			err = mail.MsgSetBodyFromFile(msg, p)
			if err != nil {
				return err
			}

			err = sendMsg(args.RootArgs.Config, msg)
			if err != nil {
				return err
			}
			return nil
		},
	}
	cmd.Flags().
		BoolVar(
			&direct, "direct-maints", direct,
			"Only send to directly affected maintainers."+
				" Equivalent to orphans.direct-maints-only in config.",
		)
	return cmd
}

func writeTemplate(outdir string, user string, o *common.Orphans) error {
	f, err := os.Create(path.Join(outdir, fmt.Sprintf("%s.txt", user)))
	if err != nil {
		return err
	}
	defer f.Close()
	td := notifs.GetUserTemplateData(o, user)
	err = notifs.UserTemplate.Execute(f, &td)
	if err != nil {
		return err
	}
	return nil
}

// WIP
// See https://lists.fedoraproject.org/archives/list/devel@lists.fedoraproject.org/message/QD3HH77G2TBXAOTMLN2LMN6W453REEGB/
func oNotifications() *cobra.Command {
	outdir := "notifs-rendered"
	cmd := &cobra.Command{
		Use:     "notifications",
		Aliases: []string{"notifs"},
		Short:   "WIP command to send individual notifications",
		RunE: func(cmd *cobra.Command, a []string) error {
			args := cmd.Context().Value(orphansArgsKey).(*OrphansArgs)
			err := os.MkdirAll(outdir, 0o755)
			if err != nil {
				return err
			}
			o, err := args.OrphansData()
			if err != nil {
				return err
			}

			// TODO: Actually send messages instead of writing to files.
			// f, err := args.RootArgs.FASCache()
			// if err != nil {
			// 	return err

			for user := range o.AllAffectedPeople {
				if user[0] == '@' {
					continue
				}
				err = writeTemplate(outdir, user, o)
				if err != nil {
					return err
				}
			}

			return nil
		},
	}
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
