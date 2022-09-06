package config

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type Config struct {
	DatabaseURL     string  `mapstructure:"DATABASE_URL"`
	Port            string  `mapstructure:"PORT"`
	BotToken        string  `mapstructure:"BOT_TOKEN"`
	Env             string  `mapstructure:"ENV"`
	DownloadDir     string  `mapstructure:"ARIA_DOWNLOAD_LOCATION"`
	SudoUsers       []int64 `mapstructure:"SUDO_USERS"`
	AuthorizedChats []int64 `mapstructure:"AUTHORIZED_CHATS"`
}

var Conf *Config

func Load() {
	// tell viper where our config file is
	viper.AddConfigPath(".")
	viper.SetConfigName("config.toml")
	viper.SetConfigType("toml")

	// override values that it has read from config file with the values of the corresponding environment variables if they exist
	viper.AutomaticEnv()

	// set defaults
	viper.SetDefault("PORT", "3000")
	viper.SetDefault("DATABASE_URL", "postgres://postgres:password@localhost:5432/katbox?sslmode=disable")
	viper.SetDefault("S3_BUCKET_NAME", "katbox")
	viper.SetDefault("ENV", "dev")

	// read in config values
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Error loading config")
	}

	// unmarshal config to struct
	Conf = &Config{}
	err = viper.Unmarshal(Conf)
	if err != nil {
		log.Fatal().Err(err).Msg("Error loading config")
	}

	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// enable debug logging if in dev
	if Conf.Env == "dev" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	log.Info().Msg("Loaded config from environment")
}
