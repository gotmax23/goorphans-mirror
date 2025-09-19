package cmds

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"go.gtmx.me/goorphans/actions"
	"go.gtmx.me/goorphans/fasjson"
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
	return cmd
}

func dgRogue() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rogue",
		Short: "Get \"rogue\" distgit group members who are not packagers",
		RunE: func(cmd *cobra.Command, argv []string) error {
			args := cmd.Context().Value(distgitArgsKey).(*DistgitArgs)
			db, err := Fas2emailCache()
			if err != nil {
				return err
			}
			f, err := fasjson.OpenCacheDB(db, fasjson.DefaultTTL)
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
