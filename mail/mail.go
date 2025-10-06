package mail

import (
	"crypto/tls"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	gomail "github.com/wneessen/go-mail"
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
	}
	if config.Secure == "tls" {
		options = append(options, gomail.WithSSL())
	}
	if config.InsecureSkipVerify {
		options = append(options, gomail.WithTLSConfig(&tls.Config{InsecureSkipVerify: true}))
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

// func (c *OurClient) SendAllWithContext(ctx context.Context, msgs ...*gomail.Msg) error {
// 	var allerr error
// 	adderrn := func(e error, n int) {
// 		if e == nil {
// 			return
// 		}
// 		allerr = errors.Join(allerr, fmt.Errorf("failed to finalize Msg at index %v: %w", n, e))
// 	}
// 	for i, msg := range msgs {
// 		adderrn(c.FinalizeMsg(msg), i)
// 	}
// 	if allerr != nil {
// 		return allerr
// 	}
// 	return c.Client.DialAndSendWithContext(ctx, msgs...)
// }
