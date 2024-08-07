package session

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// A Session represents a connection to the Discord API.
type Session struct {
	*discordgo.Session
	owner bool
}

// New creates a new Discord session and will automate some startup
// tasks if given enough information to do so.
func New(token string) (*Session, error) {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("discord: create session: %w", err)
	}

	session.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAll)

	if err := session.Open(); err != nil {
		return nil, fmt.Errorf("discord: open connection: %w", err)
	}

	return &Session{session, true}, nil
}

// Close closes a websocket and stops all listening/heartbeat goroutines.
func (s *Session) Close() error {
	if s.owner {
		return s.Session.Close()
	}

	return nil
}
