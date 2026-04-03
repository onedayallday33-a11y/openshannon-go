package config

import (
	"os"
	"strings"

	"github.com/spf13/viper"
)

func InitConfig() {
	viper.SetEnvPrefix("")
	viper.AutomaticEnv() // Bind directly to OS Env

	// Fallback Default Value
	viper.SetDefault("OPENAI_BASE_URL", "https://api.openai.com/v1")
	viper.SetDefault("OPENAI_MODEL", "gpt-4o")
}

func UseOpenAI() bool {
	val := viper.GetString("CLAUDE_CODE_USE_OPENAI")
	return val == "1" || val == "true"
}

func OpenAIApiKey() string {
	return viper.GetString("OPENAI_API_KEY")
}

func OpenAIBaseURL() string {
	return strings.TrimSuffix(viper.GetString("OPENAI_BASE_URL"), "/")
}

func OpenAIModel() string {
	return viper.GetString("OPENAI_MODEL")
}

// IsEnvTruthy is a generic helper function to check generic OS Env
func IsEnvTruthy(key string) bool {
	val := os.Getenv(key)
	return val == "1" || val == "true"
}
