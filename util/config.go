package util

import "github.com/spf13/viper"

type Config struct {
	Environment string
	ServerPort  int
	DSN         string `mapstructure:"DB_DSN"`
}

func LoadConfig(configFile string) (Config, error) {
	var conf Config

	viper.SetEnvPrefix("greenlight")
	viper.SetConfigFile(configFile)

	err := viper.ReadInConfig()
	if err != nil {
		return conf, err
	}

	viper.AutomaticEnv()

	err = viper.Unmarshal(&conf)
	if err != nil {
		return conf, err
	}

	return conf, nil
}
