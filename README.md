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
the value of `--config`) or with `GOORPHANS_<SECTION>_<OPTION>` environment
variables.

Run `goorphans dump-config` to show the default config file:

```toml
[smtp]
host = ''
port = 0
username = ''
# password can be specified in plain text or via password-cmd.
password = ''
# password-cmd can be either a string or []string of arguments.
# The first line of the command output is used as the password
# password-cmd = [""]
# From: header
from = ''
# "tls" or "starttls" (determined based on port by default)
secure = ''

[fasjson]
# Cache TTL in seconds
ttl = 604800.0
# Defaults to https://pkg.go.dev/os#UserCacheDir + "/goorphans/fasjson.db"
db = '/home/gotmax/.cache/goorphans/fasjson.db'

[orphans]
# Base url where orphans.json and orphans.txt are stored
baseurl = 'https://a.gtmx.me/orphans/'
# Whether to always re-download orphans data
download = false
```
