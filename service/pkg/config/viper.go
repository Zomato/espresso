package config

import (
	"log"
	"os"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var (
	configInstance *Config
	configMu       sync.RWMutex
	watchStarted   bool
)

func InitConfig() {
	configMu.RLock()
	if configInstance != nil {
		configMu.RUnlock()
		return
	}
	configMu.RUnlock()

	viper.SetConfigName("espressoconfig") // File name without extension
	viper.SetConfigType("yaml")           // File type

	if customPath := os.Getenv("ESPRESSO_CONFIG_PATH"); customPath != "" {
		viper.AddConfigPath(customPath)
	}

	// Search paths relative to where the binary runs in container
	viper.AddConfigPath("/app/espresso/configs") // Main config path in container
	viper.AddConfigPath("../../configs")         // For local development
	viper.AddConfigPath("./configs")             // Fallback path for local development

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		log.Fatalf("Error unmarshalling config: %v", err)
	}

	configMu.Lock()
	configInstance = cfg
	configMu.Unlock()

	startConfigWatcher()
}

func GetConfig() *Config {
	configMu.RLock()
	if configInstance == nil {
		configMu.RUnlock()
		InitConfig()
		configMu.RLock()
	}

	cfg := configInstance
	configMu.RUnlock()

	return cfg
}

func startConfigWatcher() {
	configMu.Lock()
	if watchStarted {
		configMu.Unlock()
		return
	}
	watchStarted = true
	configMu.Unlock()

	viper.OnConfigChange(func(event fsnotify.Event) {
		cfg := &Config{}
		if err := viper.Unmarshal(cfg); err != nil {
			log.Printf("Error unmarshalling updated config from %s: %v", event.Name, err)
			return
		}

		configMu.Lock()
		configInstance = cfg
		configMu.Unlock()

		log.Printf("Config reloaded from %s", event.Name)
	})

	viper.WatchConfig()
}
