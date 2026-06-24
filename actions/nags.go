package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"net/mail"
	"os"

	gomail "github.com/wneessen/go-mail"
	"go.gtmx.me/goorphans/config"
	ourmail "go.gtmx.me/goorphans/mail"
	"go.gtmx.me/goorphans/templates"
)

type TokenlessUser struct {
	User  string
	Email string
}

var TwoFANagTemplate = templates.Templates.Lookup("2fa-nag.gotmpl")

func get2FANagMsgs(
	config *config.Config,
	data []TokenlessUser,
) (msgs []*gomail.Msg, err error) {
	for _, tu := range data {
		msg := gomail.NewMsg(gomail.WithNoDefaultUserAgent())
		msg.Subject(
			fmt.Sprintf(
				"%s: ACTION REQUIRED: Two-factor authentication required for provenpackager members",
				tu.User,
			),
		)
		msg.ToMailAddress(&mail.Address{Name: tu.User, Address: tu.Email})
		if config.Nags.ReplyTo != "" {
			err = msg.ReplyTo(config.Nags.ReplyTo)
			if err != nil {
				return msgs, err
			}
		}
		err = msg.SetBodyTextTemplate(TwoFANagTemplate, &tu)
		if err != nil {
			return msgs, fmt.Errorf("failed to render template for %s: %w", tu.User, err)
		}
		msgs = append(msgs, msg)
	}
	return msgs, nil
}

func Send2FANag(ctx context.Context, config *config.Config, dataPath string) error {
	f, err := os.Open(dataPath)
	if err != nil {
		return fmt.Errorf("failed to read 2FA nag data file: %w", err)
	}
	defer f.Close()
	var data []TokenlessUser
	err = json.NewDecoder(f).Decode(&data)
	if err != nil {
		return fmt.Errorf("failed to decode 2FA nag data file: %w", err)
	}
	msgs, err := get2FANagMsgs(config, data)
	if err != nil {
		return err
	}
	return ourmail.SendMsg(ctx, config, msgs...)
}
