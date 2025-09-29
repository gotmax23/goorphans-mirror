package config

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/caarlos0/env/v11"
	"github.com/pelletier/go-toml/v2"
	"go.gtmx.me/goorphans/common"
	"go.gtmx.me/goorphans/fasjson"
)

const DefaultSentinel = "--DEFAULT--"

type Config struct {
	SMTP    SMTPConfig    `toml:"smtp"    envPrefix:"SMTP_"`
	FASJSON FASJSONConfig `toml:"fasjson" envPrefix:"FASJSON_"`
	Orphans OrphansConfig `toml:"orphans" envPrefix:"ORPHANS_"`
	// CacheDir string
}

type FASJSONConfig struct {
	TTL float64 `toml:"ttl" env:"TTL"`
	DB  string  `toml:"db"  env:"DB"`
}

type OrphansConfig struct {
	BaseURL  string `toml:"baseurl"  env:"BASEURL"`
	Download bool   `toml:"download" env:"DOWNLOAD"`
}

func LoadConfig(p string) (*Config, error) {
	var config Config
	cacheDir, err := common.CacheDir()
	if err != nil {
		return nil, err
	}
	config.FASJSON.TTL = fasjson.DefaultTTL
	config.FASJSON.DB = path.Join(cacheDir, "fasjson.db")
	// config.CacheDir = cacheDir
	config.Orphans.BaseURL = common.OrphansBaseURL

	wasDefault := false
	if p == DefaultSentinel {
		p, err = os.UserConfigDir()
		if err != nil {
			return nil, err
		}
		p = path.Join(p, "goorphans.toml")
		wasDefault = true
	}
	if p != "" {
		f, err := os.Open(p)
		if err != nil {
			if !wasDefault || !errors.Is(err, os.ErrNotExist) {
				return nil, err
			}
		}
		err = toml.NewDecoder(f).Decode(&config)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", p, err)
		}
	}
	err = env.ParseWithOptions(&config, env.Options{Prefix: "GOORPHANS_"})
	if err != nil {
		return nil, err
	}
	return &config, nil
}
