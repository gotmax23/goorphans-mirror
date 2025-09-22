package cmds

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"

	"github.com/spf13/cobra"
	"go.gtmx.me/goorphans/actions"
	"go.gtmx.me/goorphans/common"
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
	cmd.AddCommand(dgProject())
	cmd.AddCommand(dgProjectMaints())
	return cmd
}

func dgRogue() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rogue",
		Short: "Get \"rogue\" distgit group members who are not packagers",
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
			j, err := json.MarshalIndent(r, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(j))
			return nil
		},
	}
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
	return cmd
}

func dgProjectMaints() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "maints PROJECT",
		Short:   "Print out all maintainers of a project as a single list. Groups are @-prefixed.",
		Aliases: []string{"maintainers"},
		RunE: func(cmd *cobra.Command, argv []string) error {
			args := cmd.Context().Value(distgitArgsKey).(*DistgitArgs)
			maints, err := args.Client.GetAllMaints(argv[0])
			if err != nil {
				return err
			}
			return common.WriteFileLines("", maints)
		},
		Args: ArgsWrapper(cobra.ExactArgs(1)),
	}
	return cmd
}
