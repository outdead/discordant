package discordant

// Config contains credentials for Discord server.
type Config struct {
	Token       string            `json:"token" yaml:"token"`
	Prefix      string            `json:"prefix" yaml:"prefix"`
	Safemode    bool              `json:"safemode" yaml:"safemode"`
	Channels    map[string]string `json:"channels" yaml:"channels"`
	AccessOrder []string          `json:"access_order" yaml:"access_order"`
}

// Validate checks required fields and validates for allowed values.
func (cfg *Config) Validate() error {
	if cfg.Prefix == "" {
		return ErrEmptyPrefix
	}

	return nil
}
