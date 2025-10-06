package config

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/anmitsu/go-shlex"
)

const (
	portStartTLS = 587
	portFallback = 25
	portTLS      = 465
)

type SMTPConfig struct {
	Host        string `toml:"host"         env:"HOST"`
	Port        int    `toml:"port"         env:"PORT"`
	Username    string `toml:"username"     env:"USERNAME"`
	Password    string `toml:"password"     env:"PASSWORD,unset"`
	PasswordCmd any    `toml:"password_cmd" env:"PASWORD_CMD"`
	From        string `toml:"from"         env:"FROM"`
	Secure      string `toml:"secure"       env:"SECURE"`
	// Skip TLS verification
	InsecureSkipVerify bool `toml:"insecure_skip_verify" env:"INSECURE_SKIP_VERIFY"`
}

// Validate is overcomplicated code to parse SMTPConfig and handle unset
// values.
func (s *SMTPConfig) Validate() error {
	var allerr error
	adderr := func(e error) {
		allerr = errors.Join(allerr, e)
	}
	adderrf := func(format string, a ...any) {
		adderr(fmt.Errorf(format, a...))
	}

	if s.Password == "" && s.PasswordCmd != nil {
		cmd, err := parseCmd(s.PasswordCmd, "smtp.password_cmd")
		if err == nil {
			pw, err := ExecFirstLine(cmd)
			if err != nil {
				adderrf("failed to run smtp.password-cmd: %v", err)
			}
			s.Password = pw
		} else {
			adderr(err)
		}
	}

	switch s.Secure {
	case "":
		if s.Port == portStartTLS || s.Port == portFallback {
			s.Secure = "starttls"
		} else {
			// If no port is specified, default to implicit TLS.
			// https://nostarttls.secvuln.info/
			s.Secure = "tls"
			if s.Port == 0 {
				s.Port = portTLS
			}
		}
	case "tls":
		if s.Port == 0 {
			s.Port = portTLS
		}
	case "starttls":
		if s.Port == 0 {
			s.Port = portStartTLS
		}
	default:
		adderrf("invalid s.secure value %q: must be %q or %q", s.Secure, "tls", "starttls")

	}

	var missing []string
	values := map[string]*string{
		"smtp.host":     &s.Host,
		"smtp.username": &s.Username,
		"smtp.password": &s.Password,
		"smtp.from":     &s.From,
	}
	for key, value := range values {
		if *value == "" {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		adderrf(
			"missing required configuration keys: %s",
			strings.Join(missing, "; "),
		)
	}
	return allerr
}

func parseCmd(cmd any, key string) ([]string, error) {
	switch cmd := cmd.(type) {
	case string:
		// We could also just pass the string to sh -c...
		args, err := shlex.Split(cmd, true)
		if err != nil {
			return args, fmt.Errorf("invalid command value: %q", cmd)
		}
		return args, err
	case []any:
		args := make([]string, 0, len(cmd))
		for i, v := range cmd {
			v, ok := v.(string)
			if !ok {
				return nil, fmt.Errorf("invalid type at index %v: got %T; wanted string", i, v)
			}
			args = append(args, v)
		}
		return args, nil
	}
	return nil, fmt.Errorf(
		"invalid command value type for key %s: wanted string or []string and got %T",
		key, cmd,
	)
}

// ExecFirstLine is like Exec but only prints the first line of the output.
func ExecFirstLine(args []string) (string, error) {
	stdout, err := Exec(args)
	if err != nil {
		return "", err
	}
	scanner := bufio.NewScanner(bytes.NewReader(stdout))
	scanner.Scan()
	return scanner.Text(), scanner.Err()
}

func Exec(args []string) ([]byte, error) {
	if len(args) < 1 {
		return []byte{}, fmt.Errorf("invalid command value")
	}
	c := exec.Command(args[0], args[1:]...)
	c.Stderr = os.Stderr
	return c.Output()
}
