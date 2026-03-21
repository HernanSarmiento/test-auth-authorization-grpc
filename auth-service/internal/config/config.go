package config

import "github.com/spf13/viper"

type Config struct {
	AUTH_SERVER_PORT string `mapstructure:"AUTH_SERVER_PORT"`
	USER_SERVER_PORT string `mapstructure:"USER_SERVER_PORT"`
	PrivateKeyPath   string `mapstructure:"PRIVATE_KEY_PATH"`
	PublicKeyPath    string `mapstructure:"PUBLIC_KEY_PATH"`
}

func LoadConfig() (cfg Config, err error) {
	viper.SetConfigFile(".env")

	viper.AddConfigPath(".")

	viper.AddConfigPath("../")
	viper.AddConfigPath("../../")
	viper.AutomaticEnv()

	if err = viper.ReadInConfig(); err != nil {
		return
	}

	err = viper.Unmarshal(&cfg)
	return
}
