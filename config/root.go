package config

import (
	"errors"
	"os"
	"path"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/knadh/koanf/parsers/toml/v2"
	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"go.gtmx.me/goorphans/common"
	"go.gtmx.me/goorphans/fasjson"
)

const DefaultSentinel = "--DEFAULT--"

type Config struct {
	SMTP     SMTPConfig    `koanf:"smtp"`
	FASJSON  FASJSONConfig `koanf:"fasjson"`
	CacheDir string
}

type FASJSONConfig struct {
	TTL float64 `koanf:"ttl"`
	DB  string  `koanf:"db"`
}

func LoadConfig(p string) (*Config, error) {
	cacheDir, err := common.CacheDir()
	if err != nil {
		return nil, err
	}
	k := koanf.New(".")
	err = k.Set("fasjson.ttl", fasjson.DefaultTTL)
	if err != nil {
		panic(err)
	}
	err = k.Set("fasjson.db", path.Join(cacheDir, "fasjson.db"))
	if err != nil {
		panic(err)
	}

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
		err = k.Load(file.Provider(p), toml.Parser())
		if err != nil {
			if !wasDefault || !errors.Is(err, os.ErrNotExist) {
				return nil, err
			}
		}
	}
	envprefix := "GOORPHANS_"
	err = k.Load(env.Provider(".", env.Opt{
		Prefix: envprefix,
		TransformFunc: func(k, v string) (string, any) {
			k = strings.ReplaceAll(
				strings.ToLower(strings.TrimPrefix(k, envprefix)),
				"__", ".",
			)
			return k, v
		},
	}), nil)
	if err != nil {
		return nil, err
	}
	var config Config
	err = k.UnmarshalWithConf(
		"",
		&config,
		koanf.UnmarshalConf{
			DecoderConfig: &mapstructure.DecoderConfig{
				TagName:          "koanf",
				ErrorUnused:      true,
				WeaklyTypedInput: true,
			},
		},
	)
	if err != nil {
		return nil, err
	}
	config.CacheDir = cacheDir
	return &config, nil
}
