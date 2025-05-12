package config

import (
	"github.com/Zomato/espresso/service/model"
	"github.com/spf13/viper"
)

func Load(filepath string) (model.Config, error) {
	viper.SetConfigName("espressoconfig") // File name without extension
	viper.SetConfigType("yaml")           // File type
	// Search paths relative to where the binary runs in container
	viper.AddConfigPath(filepath)        // Main config path in container
	viper.AddConfigPath("../../configs") // For local development
	viper.AddConfigPath("./configs")     // Fallback path for local development

	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return model.Config{}, err
	}

	var config model.Config
	err = viper.Unmarshal(&config)
	if err != nil {
		return model.Config{}, err
	}

	// TODO: Why are these fields set in the Dockerfile and not in the config.yaml file?
	config.AppConfig.EnableUI = viper.GetBool("ENABLE_UI")
	config.AppConfig.RodBrowserBin = viper.GetString("ROD_BROWSER_BIN")

	return config, nil
}
