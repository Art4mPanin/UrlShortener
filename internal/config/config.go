package config

import "github.com/spf13/viper"

type DBConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Database string `mapstructure:"database"`
}
type Verification struct {
	EmailSender         string `mapstructure:"email_sender"`
	EmailServerUsername string `mapstructure:"email_server_username"`
	EmailServerPassword string `mapstructure:"email_server_password"`
}
type Config struct {
	IsDebug bool `mapstructure:"is_debug"`
	Listen  struct {
		Type string `mapstructure:"type"`
		Port int    `mapstructure:"port"`
	} `mapstructure:"listen"`
	Storage      DBConfig     `mapstructure:"storage"`
	Verification Verification `mapstructure:"verification"`
}

func LoadConfig() (Config, error) {
	var config Config
	viper.AddConfigPath("config")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return config, err
	}
	if err := viper.Unmarshal(&config); err != nil {
		return config, err
	}

	return config, nil
}
