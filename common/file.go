package common

import (
	"os"
	"path"
)

func CacheDir() (string, error) {
	p, err := os.UserCacheDir()
	if err != nil {
		return p, err
	}
	p = path.Join(p, "goorphans")
	return p, os.MkdirAll(p, 0o700)
}
