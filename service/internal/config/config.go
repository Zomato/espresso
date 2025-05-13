package config

import (
	"fmt"

	"github.com/Zomato/espresso/service/model"
	"github.com/spf13/viper"
)

func Load() (model.Config, error) {
	viper.AutomaticEnv()

	viper.SetDefault(CONFIG_FILE_NAME, "espressoconfig")
	viper.SetDefault(CONFIG_FILE_TYPE, "yaml")
	viper.SetDefault(CONFIG_FILE_PATH, "/app/espresso/configs")

	viper.SetConfigName(viper.GetString(CONFIG_FILE_NAME)) // File name without extension
	viper.SetConfigType(viper.GetString(CONFIG_FILE_TYPE)) // File type

	// Search paths relative to where the binary runs in container
	viper.AddConfigPath(CONFIG_FILE_PATH) // Main config path in container
	viper.AddConfigPath("../../configs")  // For local development
	viper.AddConfigPath("./configs")      // Fallback path for local development

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
	config.AppConfig.EnableUI = viper.GetBool(ENABLE_UI)
	config.AppConfig.RodBrowserBin = viper.GetString(ROD_BROWSER_BIN)
	if config.AppConfig.RodBrowserBin == "" {
		return model.Config{}, fmt.Errorf("environment variable %s not set", ROD_BROWSER_BIN)
	}

	return config, nil
}
