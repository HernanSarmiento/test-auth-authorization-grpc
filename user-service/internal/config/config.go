package config

import "github.com/spf13/viper"

type Config struct {
	DB_PORT     string `mapstructure:"DB_PORT"`
	DB_HOST     string `mapstructure:"DB_HOST"`
	DB_USER     string `mapstructure:"DB_USER"`
	DB_PASSWORD string `mapstructure:"DB_PASSWORD"`
	DB_NAME     string `mapstructure:"DB_NAME"`
	DB_SSLMODE  string `mapstructure:"DB_SSLMODE"`
	SERVER_PORT string `mapstructure:"SERVER_PORT"`
}

func LoadConfig() (cfg Config, err error) {
	viper.AddConfigPath(".")
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	viper.AddConfigPath("../..")
	viper.AutomaticEnv()

	if err = viper.ReadInConfig(); err != nil {
		return
	}

	err = viper.Unmarshal(&cfg)
	return
}
