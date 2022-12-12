package config

import (
	"fmt"

	"github.com/jessevdk/go-flags"
)

// Config is a banhammer config.
type Config struct {
	LoggerLever    string `long:"logger-level" env:"LOGGER_LEVEL" description:"Logger level" default:"info"`
	HeuristicsPath string `long:"heuristics-path" env:"HEURISTICS_PATH" description:"Path to heuristics file" default:"heuristics.toml"`

	APIToken                 string `long:"api-token:" env:"API_TOKEN" description:"VK API token" required:"true"`
	CallbackConfirmationCode string `long:"callback-confirmation-code" env:"CALLBACK_CONFIRMATION_CODE" description:"Callback confirmation code from VK" required:"true"`
	HTTPAddr                 string `long:"http-addr" env:"HTTP_ADDR" description:"HTTP server address" default:":8080"`
}

// ParseConfig parses banhammer config.
func ParseConfig() (*Config, error) {
	var cfg Config
	if _, err := flags.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse: %w", err)
	}

	return &cfg, nil
}
