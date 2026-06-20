package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	App struct {
		Name string `mapstructure:"name"`
		Port string `mapstructure:"port"`
	} `mapstructure:"app"`
	Database DbConfig    `mapstructure:"db"`
	Grpc     GrpcConfig  `mapstructure:"grpc"`
	Kafka    KafkaConfig `mapstructure:"kafka"`
}

type DbConfig struct {
	Addr struct {
		Host string `mapstructure:"host"`
		Port string `mapstructure:"port"`
	} `mapstructure:"addr"`
	Username        string `mapstructure:"username"`
	Password        string `mapstructure:"password"`
	Dbname          string `mapstructure:"dbname"`
	AuthSource      string `mapstructure:"authsource"`
	MaxConns        int    `mapstructure:"max_conns"`
	MinConns        int    `mapstructure:"min_conns"`
	MaxConnIdleTime int    `mapstructure:"maxconns_idletime"`
}
type GrpcConfig struct {
	AuthService struct {
		Host string `mapstructure:"host"`
		Port string `mapstructure:"port"`
	} `mapstructure:"authservice"`
}
type KafkaConfig struct {
	Addr struct {
		Host string `mapstructure:"host"`
		Port string `mapstructure:"port"`
	} `mapstructure:"addr"`
	Topic      string `mapstructure:"topic"`
	AsyncWrite bool   `mapstructure:"asyncwrite"`
}

func (d DbConfig) DSN() string {
	return fmt.Sprintf(
		"mongodb://%s:%s@%s:%s/%s?authSource=%s",
		d.Username, d.Password, d.Addr.Host, d.Addr.Port, d.Dbname, d.AuthSource,
	)
}
func Load() (*Config, error) {
	v := viper.New()

	// set defaults
	v.SetDefault("app.name", "user-service")
	v.SetDefault("app.port", "8000")
	v.SetDefault("db.addr.host", "127.0.0.1")
	v.SetDefault("db.addr.port", "27017")
	v.SetDefault("kafka.topic", "user-service")
	v.SetDefault("kafka.asyncwrite", "false")

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// config: env
	v.SetConfigName("dev")
	v.SetConfigType(".env")
	v.AddConfigPath("../")
	if err := v.MergeInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return nil, fmt.Errorf("err reading .env: %w", err)
		}
	}

	// config: yml
	v.SetConfigName("config")
	v.SetConfigType(".yml")
	v.AddConfigPath("../")
	if err := v.MergeInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return nil, fmt.Errorf("err reading cfg.yml: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		fmt.Printf("unable to decode into struct: %v\n", err)
		return nil, err
	}

	return &cfg, nil
}
