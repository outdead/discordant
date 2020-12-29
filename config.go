package discordant

import "errors"

// Config contains credentials for Discord server.
type Config struct {
	Token    string            `json:"token" yaml:"token"`
	Prefix   string            `json:"prefix" yaml:"prefix"`
	Channels map[string]string `json:"channels" yaml:"channels"`
}

// Validate checks required fields and validates for allowed values.
func (cfg *Config) Validate() error {
	if cfg.Token == "" {
		return errors.New("field token is not defined")
	}

	if cfg.Prefix == "" {
		return errors.New("field prefix is not defined")
	}

	return nil
}
