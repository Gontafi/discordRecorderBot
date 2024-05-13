package pkg

import (
	"github.com/spf13/viper"
)

var Config *Credentials

type Credentials struct {
	DiscordBotToken string `mapstructure:"bot_token"`
	AccessKeyS3     string `mapstructure:"aws_access_key"`
	SecretKeyS3     string `mapstructure:"aws_secret_key"`
}

func ReadCredentials() error {
	Config = new(Credentials)

	viper.SetConfigName("config") // name of config file (without extension)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	err := viper.Unmarshal(&Config)
	if err != nil {
		return err
	}

	return nil
}
