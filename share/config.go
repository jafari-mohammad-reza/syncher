package share

import (
	"errors"
	"github.com/spf13/viper"
	"log"
)

type MinIO struct {
	Endpoint        string `mapstructure:"MINIO_ENDPOINT"`
	AccessKeyID     string `mapstructure:"MINIO_ACCESS_KEY_ID"`
	SecretAccessKey string `mapstructure:"MINIO_SECRET_ACCESS`
	UseSSL          bool   `mapstructure:"MINIO_USE_SSL"`
}
type ServerConfig struct {
	NatsUrl  string `mapstructure:"NATS_URL"`
	ServerId string
	MinIO
}
type ClientConfig struct {
	NatsUrl  string `mapstructure:"NATS_URL"`
	ClientId string
}

func GetServerConfig() (*ServerConfig, error) {
	v, err := InitConfig("server.yaml")
	if err != nil {
		return nil, err
	}
	var cfg ServerConfig
	err = v.Unmarshal(&cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func GetClientConfig() (*ClientConfig, error) {
	v, err := InitConfig("client.yaml")
	if err != nil {
		return nil, err
	}
	var cfg ClientConfig
	err = v.Unmarshal(&cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
func InitConfig(name string) (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigType("yml")
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
	return v, nil
}
