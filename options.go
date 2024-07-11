package discordant

import (
	"github.com/bwmarrin/discordgo"
	"github.com/outdead/discordant/internal/session"
)

// Option can be used to create a customized connections.
type Option func(d *Discordant)

// SetSession sets discordgo session to Discordant.
func SetSession(ses *discordgo.Session) Option {
	return func(d *Discordant) {
		d.session = &session.Session{Session: ses}
	}
}

// SetLogger sets logger to Discordant.
func SetLogger(logger Logger) Option {
	return func(d *Discordant) {
		d.logger = logger
	}
}
