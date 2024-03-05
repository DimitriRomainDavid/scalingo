package config

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type Config struct {
	GinMode string

	GitHubToken            string
	GitHubCredentials      bool
	GitHubURL              string
	GitHubVersion          string
	LatestCreatedRepoRetry int

	OutputSize          int
	ProcessingBatchSize int

	HTTPPort    string
	HTTPAddress string
}

func ProvideConfig() *Config {
	initDefault()
	viper.SetConfigFile(".env")
	if err := viper.ReadInConfig(); err != nil {
		panic("Error unmarshalling config: " + err.Error())
	}

	return &Config{
		GinMode: viper.GetString("GIN_MODE"),

		GitHubToken:            viper.GetString("GITHUB_TOKEN"),
		GitHubCredentials:      viper.GetBool("GITHUB_CREDENTIALS"),
		GitHubURL:              viper.GetString("GITHUB_URL"),
		GitHubVersion:          viper.GetString("GITHUB_VERSION"),
		LatestCreatedRepoRetry: viper.GetInt("LATEST_CREATED_REPO_RETRY"),

		OutputSize:          viper.GetInt("OUTPUT_SIZE"),
		ProcessingBatchSize: viper.GetInt("PROCESSING_BATCH_SIZE"),

		HTTPPort:    viper.GetString("HTTP_PORT"),
		HTTPAddress: viper.GetString("HTTP_ADDRESS"),
	}
}

func (c *Config) GetServerAddress() string {
	return c.HTTPAddress + ":" + c.HTTPPort
}

func initDefault() {
	viper.SetDefault("GIN_MODE", gin.DebugMode)

	viper.SetDefault("GITHUB_TOKEN", "")
	viper.SetDefault("GITHUB_CREDENTIALS", false)
	viper.SetDefault("GITHUB_URL", "")
	viper.SetDefault("GITHUB_VERSION", "")

	viper.SetDefault("HTTP_PORT", 5000) //nolint: gomnd
	viper.SetDefault("HTTP_ADDRESS", "")
}
