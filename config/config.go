package config

import (
	"time"

	"github.com/spf13/viper"
	"github.com/vatsimnerd/simwatch-providers/ourairports"
	vatsimapi "github.com/vatsimnerd/simwatch-providers/vatsim-api"
	vatspydata "github.com/vatsimnerd/simwatch-providers/vatspy-data"
)

type WebConfig struct {
	Addr string `mapstructure:"addr,omitempty"`
	CORS bool   `mapstructure:"cors,omitempty"`
}

type TrackConfigOptions struct {
	Addr        string        `mapstructure:"addr,omitempty"`
	Password    string        `mapstructure:"password,omitempty"`
	DB          int           `mapstructure:"db,omitempty"`
	PurgePeriod time.Duration `mapstructure:"purge_period,omitempty"`
}

type TrackConfig struct {
	Engine  string             `mapstructure:"engine,omitempty"`
	Options TrackConfigOptions `mapstructure:"options,omitempty"`
}

type Config struct {
	API      vatsimapi.Config   `mapstructure:"api,omitempty"`
	Data     vatspydata.Config  `mapstructure:"data,omitempty"`
	Runways  ourairports.Config `mapstructure:"runways,omitempty"`
	LogLevel string             `mapstructure:"log_level,omitempty"`
	Web      WebConfig          `mapstructure:"web,omitempty"`
	Track    TrackConfig        `mapstructure:"track,omitempty"`
}

func Read(filename string) (*Config, error) {
	var err error

	cfg := new(Config)
	viper.SetConfigName(filename)
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/simwatch")

	viper.SetDefault("api.url", vatsimapi.VatsimAPIURL)
	viper.SetDefault("api.poll.period", 15*time.Second)
	viper.SetDefault("api.poll.timeout", 3*time.Second)
	viper.SetDefault("api.boot.retries", 5)
	viper.SetDefault("api.boot.retry_cooldown", 3*time.Second)

	viper.SetDefault("data.data_url", vatspydata.DefaultDataURL)
	viper.SetDefault("data.boundaries_url", vatspydata.DefaultBoundariesURL)
	viper.SetDefault("data.poll.period", 24*time.Hour)
	viper.SetDefault("data.poll.timeout", 3*time.Second)
	viper.SetDefault("data.boot.retries", 5)
	viper.SetDefault("data.boot.retry_cooldown", 3*time.Second)

	viper.SetDefault("runways.url", ourairports.OurairportsRunwaysURL)
	viper.SetDefault("runways.poll.period", 24*time.Hour)
	viper.SetDefault("runways.poll.timeout", 3*time.Second)
	viper.SetDefault("runways.boot.retries", 5)
	viper.SetDefault("runways.boot.retry_cooldown", 3*time.Second)

	viper.SetDefault("web.addr", "localhost:5000")
	viper.SetDefault("web.cors", false)

	viper.SetDefault("track.engine", "memory")
	viper.SetDefault("track.options.purge_period", "24h")
	viper.SetDefault("track.options.addr", "localhost:6379")
	viper.SetDefault("track.options.password", "")
	viper.SetDefault("track.options.db", 0)

	err = viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	err = viper.Unmarshal(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
