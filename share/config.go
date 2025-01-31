package share

import (
	"errors"
	"log"
	"log/slog"

	"github.com/spf13/viper"
)

type Config struct {
	AppType     string `mapstructure:"APP_TYPE"`
	NatsUrl     string `mapstructure:"NATS_URL"`
	PostgresUrl string `mapstructure:"POSTGRES_URL"`
}

func InitConfig(name string) (*Config, error) {
	v := viper.New()
	v.SetConfigType("dotenv")
	v.SetConfigName(name)
	v.AddConfigPath(".")
	v.AutomaticEnv()
	v.WatchConfig()
	err := v.ReadInConfig()
	if err != nil {
		log.Printf("Unable to read config: %v", err)
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			return nil, errors.New("config file not found")
		}
		return nil, err
	}
	var cfg Config
	err = v.Unmarshal(&cfg)
	if err != nil {
		slog.Error("Unable to read config: %v", err)
		return nil, err
	}
	return &cfg, nil

}
