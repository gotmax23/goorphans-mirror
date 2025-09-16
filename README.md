# Goorphans

[![builds.sr.ht status](https://builds.sr.ht/~gotmax23/goorphans/commits/main.svg)](https://builds.sr.ht/~gotmax23/goorphans/commits/main?)

WIP tooling for managing the orphaned packages process and other Fedora
releated things.

```fish
sudo dnf install -y golang
# Or whatever your shell supports...
fish_add_path $(go env GOPATH)/bin
go install -v go.gtmx.me/goorphans@latest
goorphans --help
```
