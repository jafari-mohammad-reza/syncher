package share

import (
	"errors"
	"log"

	"github.com/spf13/viper"
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
	NatsUrl      string `mapstructure:"NATS_URL"`
	ClientId     string `mapstructure:"CLIENT_ID"`
	HttpPort     string   `mapstructure:"HTTP_PORT"`
	SyncDirs     []string `mapstructure:"SYNC_DIRS"`
	SyncInterval int      `mapstructure:"SYNC_INTERVAL"`
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

var clientViper *viper.Viper

func InitClientConfig() {
	v, err := InitConfig("client.yaml")
	if err != nil {
		panic(err)
	}
	clientViper = v
}
func GetClientConfig() (*ClientConfig, error) {
	var cfg ClientConfig
	err := clientViper.Unmarshal(&cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
func WriteClientConfig() {
	clientViper.WriteConfig()
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
