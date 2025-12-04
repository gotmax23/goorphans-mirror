package cmds

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"net/url"
	"os"
	"slices"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/spf13/cobra"
	"go.gtmx.me/goorphans/actions"
	"go.gtmx.me/goorphans/common"
	"go.gtmx.me/goorphans/distgit"
	"go.gtmx.me/goorphans/pagure"
)

const DefaultDistgitURL = "https://src.fedoraproject.org"

var distgitArgsKey = &argsKeyType{"fas2email"}

type DistgitArgs struct {
	Client *pagure.Client
}

func NewDistgitCmd() *cobra.Command {
	var distgit string
	var args DistgitArgs
	cmd := &cobra.Command{
		Use:     "distgit",
		Short:   "Commands for working with Fedora distgit",
		Aliases: []string{"dg"},
		PersistentPreRunE: func(cmd *cobra.Command, argv []string) error {
			u, err := url.Parse(distgit)
			if err != nil {
				return err
			}
			args.Client = pagure.NewClient(u, nil)
			cmd.SetContext(context.WithValue(cmd.Context(), distgitArgsKey, &args))
			return nil
		},
	}
	cmd.PersistentFlags().StringVar(&distgit, "distgit", DefaultDistgitURL, "distgit URL")
	cmd.AddCommand(dgRogue())
	cmd.AddCommand(dgRogueMaints())
	cmd.AddCommand(dgRogueOrphans())
	cmd.AddCommand(dgProject())
	cmd.AddCommand(dgMaintEmails())
	return cmd
}

func dgRogue() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rogue-members",
		Aliases: []string{"rogue"},
		Short:   "Get \"rogue\" distgit group members who are not packagers",
		RunE: func(cmd *cobra.Command, argv []string) error {
			args := cmd.Context().Value(distgitArgsKey).(*DistgitArgs)
			rargs := cmd.Context().Value(rootArgsKey).(*RootArgs)
			f, err := rargs.FASCache()
			if err != nil {
				return err
			}
			r, err := actions.GetRoguePackagerGroupMembers(f, args.Client)
			if err != nil {
				return err
			}
			// TODO: Print results as JSON for now
			return JSONToStdout(r)
		},
	}
	return cmd
}

func dgRogueMaints() *cobra.Command {
	namesOnly := false
	cmd := &cobra.Command{
		Use:   "rogue-maints",
		Short: "Get \"rogue\" package admins who are not packagers",
		RunE: func(cmd *cobra.Command, argv []string) error {
			// args := cmd.Context().Value(distgitArgsKey).(*DistgitArgs)
			rargs := cmd.Context().Value(rootArgsKey).(*RootArgs)
			f, err := rargs.FASCache()
			if err != nil {
				return err
			}
			e := distgit.NewExtrasClient(rargs.HTTPClient)
			r, err := actions.GetRoguePackageAdmins(f, e)
			if err != nil {
				return err
			}
			slices.SortFunc(r, func(a, b actions.RoguePackageAdmin) int {
				return strings.Compare(a.Package, b.Package)
			})
			for _, m := range r {
				if namesOnly {
					fmt.Println(m.Package)
				} else {
					fmt.Printf("%s %s\n", m.Package, m.Admin)
				}
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&namesOnly, "names", "N", namesOnly, "Only print package names")
	return cmd
}

func dgRogueOrphans() *cobra.Command {
	namesOnly := false
	cmd := &cobra.Command{
		Use:   "rogue-orphans",
		Short: "Get packagers that have real admins and orphan assignees",
		RunE: func(cmd *cobra.Command, argv []string) error {
			// args := cmd.Context().Value(distgitArgsKey).(*DistgitArgs)
			rargs := cmd.Context().Value(rootArgsKey).(*RootArgs)
			e := distgit.NewExtrasClient(rargs.HTTPClient)
			poc, err := e.GetPagurePOC()
			if err != nil {
				return err
			}
			r := map[string]distgit.ExtrasPagurePOCTypes{}
			for name, pocs := range poc.RPMS {
				if pocs.Admin != common.OrphanUID && pocs.Fedora == common.OrphanUID {
					isretired, err := e.IsRetired(name, "rawhide")
					if err != nil {
						return fmt.Errorf(
							"failed to check if %s is retired: %w",
							name,
							err,
						)
					}
					if !isretired {
						r[name] = pocs
					}
				}
			}
			names := slices.Sorted(maps.Keys(r))
			for _, name := range names {
				pocs := r[name]
				if namesOnly {
					fmt.Println(name)
				} else {
					fmt.Printf("%s admin=%s fedora=%s epel=%s\n", name, pocs.Admin, pocs.Fedora, pocs.EPEL)
				}
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&namesOnly, "names", "N", namesOnly, "Only print package names")
	// cmd.Flags().BoolVarP()
	return cmd
}

func dgMaintEmails() *cobra.Command {
	prefix := "rpms/"
	noGroups := false
	cmd := &cobra.Command{
		Use:   "maint-emails PACKAGE...",
		Short: "Get package maintainer emails",
		RunE: func(cmd *cobra.Command, argv []string) error {
			args := cmd.Context().Value(distgitArgsKey).(*DistgitArgs)
			rargs := cmd.Context().Value(rootArgsKey).(*RootArgs)
			f, err := rargs.FASCache()
			if err != nil {
				return err
			}
			s := mapset.NewThreadUnsafeSet[string]()
			for _, p := range argv {
				m, err := args.Client.GetAllMaints(prefix+p, !noGroups)
				if err != nil {
					return err
				}
				s.Append(m...)
			}
			mails, err := f.GetIterEmailsMap(mapset.Elements(s))
			if err != nil {
				return err
			}
			return common.WriteFileLines("-", slices.Sorted(maps.Values(mails)))
		},
		Args: ArgsWrapper(cobra.MinimumNArgs(1)),
	}
	cmd.Flags().StringVar(&prefix, "prefix", prefix, "Pagure project prefix")
	cmd.Flags().BoolVarP(&noGroups, "no-groups", "G", noGroups, "Don't expand groups")
	return cmd
}

func JSONToStdout(data any) error {
	w := bufio.NewWriter(os.Stdout)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	err := enc.Encode(data)
	if err != nil {
		return err
	}
	return w.Flush()
}

func dgProject() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "project",
		Short:   "Generic commands for dealing with Pagure projects",
		Aliases: []string{"p"},
	}
	cmd.AddCommand(dgProjectInfo())
	cmd.AddCommand(dgProjectMaints())
	return cmd
}

func dgProjectInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use: "info PROJECT",
		RunE: func(cmd *cobra.Command, argv []string) error {
			args := cmd.Context().Value(distgitArgsKey).(*DistgitArgs)
			data, err := args.Client.GetProject(argv[0])
			if err != nil {
				return err
			}
			return JSONToStdout(data)
		},
		Args: ArgsWrapper(cobra.ExactArgs(1)),
	}
	cmd.AddCommand(dgProjectMaints())
	return cmd
}

func dgProjectMaints() *cobra.Command {
	noGroups := false
	cmd := &cobra.Command{
		Use:     "maints PROJECT",
		Short:   "Print out all maintainers of a project as a single list. Groups are @-prefixed.",
		Aliases: []string{"maintainers"},
		RunE: func(cmd *cobra.Command, argv []string) error {
			args := cmd.Context().Value(distgitArgsKey).(*DistgitArgs)
			maints, err := args.Client.GetAllMaints(argv[0], !noGroups)
			if err != nil {
				return err
			}
			slices.Sort(maints)
			return common.WriteFileLines("-", maints)
		},
		Args: ArgsWrapper(cobra.ExactArgs(1)),
	}
	cmd.Flags().BoolVarP(&noGroups, "no-groups", "G", noGroups, "Don't expand groups")
	return cmd
}
