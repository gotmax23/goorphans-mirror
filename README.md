# Goorphans

[![builds.sr.ht status](https://builds.sr.ht/~gotmax23/goorphans/commits/main.svg)](https://builds.sr.ht/~gotmax23/goorphans/commits/main?)

WIP tooling for managing the orphaned packages process and other Fedora
releated things.

## Installation

```fish
# krb5 is needed for the FASJSON bindings
sudo dnf install -y golang krb5-devel
# Or whatever your shell supports...
fish_add_path $(go env GOPATH)/bin
go install -v go.gtmx.me/goorphans@latest
goorphans --help
```

## Configuration

goorphans can be configured with a TOML file at `~/.config/goorphans.toml` (or
the value of `--config`) or with environment variables.

Run `goorphans dump-config` to show the default config file.
The code block below shows all the supported options.

```toml
[smtp]
# Env: GOORPHANS_SMTP_HOST
host = ''
# Env: GOORPHANS_SMTP_PORT
port = 0
# Env: GOORPHANS_SMTP_USERNAME
username = ''
# Env: GOORPHANS_SMTP_PASSWORD
# password can be specified in plain text or via password-cmd.
password = ''
# Env: GOORPHANS_SMTP_PASSWORD_CMD
# password-cmd can be either a string or []string of arguments.
# The first line of the command output is used as the password
password-cmd = [""]
# Env: GOORPHANS_SMTP_FROM
# From: header
from = ''
# Env: GOORPHANS_SMTP_SECURE
# "tls" or "starttls" (determined based on port by default)
secure = ''
# Env: GOORPHANS_SMTP_INSECURE_SKIP_VERIFY
# Don't validate SMTP server TLS certificates.
insecure-skip-verify = false

[fasjson]
# Env: GOORPHANS_FASJSON_TTL
# Cache TTL in seconds
ttl = 604800.0
# Env: GOORPHANS_FASJSON_DB
# Defaults to https://pkg.go.dev/os#UserCacheDir + "/goorphans/fasjson.db"
db = '/home/gotmax/.cache/goorphans/fasjson.db'

[orphans]
# Env: GOORPHANS_ORPHANS_BASEURL
# Base url where orphans.json and orphans.txt are stored
baseurl = 'https://a.gtmx.me/orphans/'
# Env: GOORPHANS_ORPHANS_DOWNLOAD
# Whether to always re-download orphans data
download = false
# Env: GOORPHANS_ORPHANS_TO
# To for the Orphaned Packages report
to = ['devel-announce@lists.fedoraproject.org']
# Env: GOORPHANS_ORPHANS_REPLY_TO
# To for the Orphaned Packages report
reply-to = 'devel@lists.fedoraproject.org'
# Env: GOORPHANS_ORPHANS_BCC
# Can be used to send a copy to yourself.
bcc = []
# Env: GOORPHANS_ORPHANS_DIRECT_MAINTS_ONLY
direct-maints-only = false
```
