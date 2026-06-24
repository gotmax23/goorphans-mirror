package mail

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"

	gomail "github.com/wneessen/go-mail"
	"github.com/wneessen/go-mail/smtp"
	"go.gtmx.me/goorphans/config"
)

func NewClient(config *config.SMTPConfig, opts ...gomail.Option) (*gomail.Client, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	options := []gomail.Option{
		gomail.WithPort(config.Port),
		gomail.WithSMTPAuth(gomail.SMTPAuthPlain),
		gomail.WithUsername(config.Username),
		gomail.WithPassword(config.Password),
		gomail.WithPort(config.Port),
		gomail.WithTimeout(time.Minute),
	}
	if config.Secure == "tls" {
		options = append(options, gomail.WithSSL())
	}
	if config.InsecureSkipVerify {
		options = append(
			options,
			gomail.WithTLSConfig(&tls.Config{InsecureSkipVerify: true}),
		)
	}
	c, err := gomail.NewClient(config.Host, options...)
	return c, err
}

func getMsgID(from string) (string, error) {
	_, after, found := strings.Cut(from, "@")
	if !found {
		// This shouldn't happen; go-mail already validated the address.
		return "", fmt.Errorf("invalid From address")
	}
	// This Message-ID generator is more robust and doesn't hard
	// code the local user's hostname like go-mail's implementation does.
	return GenerateMessageIDWithHostname(after)
}

func FinalizeMsg(config *config.SMTPConfig, msg *gomail.Msg) error {
	err := msg.From(config.From)
	if err != nil {
		return err
	}
	// Get the parsed From value
	from := msg.GetFrom()
	msgid, err := getMsgID(from[0].Address)
	if err != nil {
		return err
	}
	msg.SetMessageIDWithValue(msgid)
	msg.SetDateWithValue(time.Now().UTC())
	return nil
}

func MsgSetBodyFromFile(msg *gomail.Msg, name string) error {
	msg.SetBodyWriter(gomail.TypeTextPlain, func(w io.Writer) (int64, error) {
		f, err := os.Open(name)
		if err != nil {
			return 0, err
		}
		defer f.Close()
		return io.Copy(w, f)
	})
	return nil
}

func SendMsg(ctx context.Context, config *config.Config, msgs ...*gomail.Msg) error {
	var c *gomail.Client
	var sclient *smtp.Client
	var err error
	if config.SMTP.OutgoingDir == "" {
		c, err = NewClient(&config.SMTP)
		if err != nil {
			return err
		}
		sclient, err = c.DialToSMTPClientWithContext(ctx)
		if err != nil {
			return err
		}
		defer c.CloseWithSMTPClient(sclient)
	}
	// FinalizeMsgs before we begin sending in case there's an error.
	for i, msg := range msgs {
		err = FinalizeMsg(&config.SMTP, msg)
		if err != nil {
			return fmt.Errorf("failed to finalize msg at index %v: %w", i, err)
		}
	}
	for i, msg := range msgs {
		err := FinalizeMsg(&config.SMTP, msg)
		if err != nil {
			return err
		}
		r, _ := msg.GetRecipients()
		fmt.Printf(
			"(%d/%d) Sending %q to %d recipients...\n",
			i+1, len(msgs), msg.GetGenHeader("Subject")[0], len(r),
		)
		if config.SMTP.OutgoingDir == "" {
			err = c.SendWithSMTPClient(sclient, msg)
		} else {
			p := path.Join(config.SMTP.OutgoingDir, msg.GetMessageID()+".eml")
			err = msg.WriteToFile(p)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
