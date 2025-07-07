package discordant

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// discordMaxMessageLenValidate max discord message length for internal validation.
const discordMaxMessageLenValidate = 1990

const (
	stateStart  = "start"
	stateQuotes = "quotes"
	stateArg    = "arg"
)

// Context is an interface represents the context of the current Discord command.
type Context interface {
	Command() *Command
	Discordant() *Discordant
	Request() *discordgo.MessageCreate
	ChannelID() string
	QueryString() string
	QueryParams() ([]string, error)
	Send(msg string, params ...string) error
}

type context struct {
	command    *Command
	discordant *Discordant
	request    *discordgo.MessageCreate
}

// Command returns received command.
func (c *context) Command() *Command {
	return c.command
}

// Discordant returns Discordant instance.
func (c *context) Discordant() *Discordant {
	return c.discordant
}

// Request returns the data for a MessageCreate event from request query.
func (c *context) Request() *discordgo.MessageCreate {
	return c.request
}

// ChannelID returns the ID of the channel in which the message was sent.
func (c *context) ChannelID() string {
	return c.request.ChannelID
}

// QueryString returns the URL query string.
func (c *context) QueryString() string {
	return c.Command().Arg
}

// QueryParams returns the query parameters as slice.
func (c *context) QueryParams() ([]string, error) { //nolint: cyclop, funlen // indivisible
	query := c.Command().Arg

	var args []string

	state := stateStart
	current := ""
	quote := "\""
	escapeNext := true

	for i, command := range query {
		if i == 0 && string(command) == `"` {
			escapeNext = false
		}

		if state == stateQuotes {
			if string(command) != quote {
				current += string(command)
			} else {
				args = append(args, current)
				current = ""
				state = stateStart
			}

			continue
		}

		if escapeNext {
			current += string(command)
			escapeNext = false

			continue
		}

		if command == '\\' {
			escapeNext = true

			continue
		}

		if command == '"' || command == '\'' {
			state = stateQuotes
			quote = string(command)

			continue
		}

		if state == stateArg {
			if command == ' ' || command == '\t' {
				args = append(args, current)
				current = ""
				state = stateStart
			} else {
				current += string(command)
			}

			continue
		}

		if command != ' ' && command != '\t' {
			state = stateArg
			current += string(command)
		}
	}

	if state == stateQuotes {
		return []string{}, fmt.Errorf("%w: %s", ErrUnclosedQuote, query)
	}

	if current != "" {
		args = append(args, current)
	}

	return args, nil
}

// Send sends message to discord channel.
// TODO: Add the option of posting to Discord channel.
// That is, what to do with messages that more than 2000 characters.
// 1. Send as file.
// 2. Send as multiple messages
// TODO: Add sending format selection.
// 1. Plaint text - As is.
// 2. Format JSON ```json\n%s\n```
// 3. Embed - beautiful.
func (c *context) Send(msg string, params ...string) error {
	// Send normal message.
	if len([]rune(msg)) <= discordMaxMessageLenValidate {
		if _, err := c.discordant.session.ChannelMessageSend(c.request.ChannelID, msg); err != nil {
			return fmt.Errorf("discordant send: %w", err)
		}

		return nil
	}

	// Message is too big. Attach as file.
	msg = strings.TrimPrefix(msg, "```json\n")
	msg = strings.TrimSuffix(msg, "\n```")

	var buf bytes.Buffer

	fileName := "message.txt"

	if len(params) > 0 {
		fileName = params[0]
	}

	if _, err := buf.Write([]byte(msg)); err != nil {
		return fmt.Errorf("discordant send: %w", err)
	}

	ms := &discordgo.MessageSend{Files: []*discordgo.File{
		{Name: fileName, Reader: bufio.NewReader(&buf)},
	}}

	if _, err := c.discordant.session.ChannelMessageSendComplex(c.request.ChannelID, ms); err != nil {
		return fmt.Errorf("discordant send: %w", err)
	}

	return nil
}
