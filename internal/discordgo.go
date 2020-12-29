package internal

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
func NewSession(token string) (*Session, error) {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("error creating discord session: %s", err)
	}

	if err := session.Open(); err != nil {
		return nil, fmt.Errorf("open connection to discord error: %s", err)
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
