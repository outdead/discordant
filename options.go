package discordant

import (
	"github.com/bwmarrin/discordgo"
	"github.com/outdead/discordant/internal"
)

// Option can be used to a create a customized connections.
type Option func(d *Discordant)

// SetSession sets discordgo session to Discordant.
func SetSession(session *discordgo.Session) Option {
	return func(d *Discordant) {
		d.session = &internal.Session{Session: session}
	}
}

// SetLogger sets logger to Discordant.
func SetLogger(logger Logger) Option {
	return func(d *Discordant) {
		d.logger = logger
	}
}
