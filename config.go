package discordant

// Config contains credentials for Discord server.
type Config struct {
	Token    string            `json:"token" yaml:"token"`
	Prefix   string            `json:"prefix" yaml:"prefix"`
	Channels map[string]string `json:"channels" yaml:"channels"`
}

// Validate checks required fields and validates for allowed values.
func (cfg *Config) Validate() error {
	if cfg.Token == "" {
		return ErrEmptyToken
	}

	if cfg.Prefix == "" {
		return ErrEmptyPrefix
	}

	return nil
}
