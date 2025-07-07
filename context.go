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

// QueryParams parses the command query string and returns a slice of arguments.
// It handles quoted arguments (both single and double quotes) and escape characters.
//
// The function implements a simple state machine to parse the query string with these rules:
// 1. Spaces separate arguments unless they're within quotes
// 2. Both single (') and double (") quotes are supported
// 3. Backslash (\) can be used to escape special characters
// 4. Unclosed quotes will return an error
//
// Returns:
//   - []string: slice of parsed arguments
//   - error: if there's an unclosed quote in the input
//
// Note: The function is marked with nolint for cyclop and funlen as the state machine
// logic is inherently complex but intentionally kept as a single unit for clarity.
func (c *context) QueryParams() ([]string, error) { //nolint: cyclop, funlen // indivisible
	query := c.Command().Arg

	var args []string

	// State machine states
	state := stateStart // Initial parsing state
	current := ""       // Current argument being built
	quote := "\""       // Type of quote we're currently in (if in quotes)
	escapeNext := true  // Whether next character should be escaped

	for i, command := range query {
		// Special case: first character is a quote (disable escaping)
		if i == 0 && string(command) == `"` {
			escapeNext = false
		}

		// When inside quotes, accept all characters until closing quote
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

		// Handle escaped characters
		if escapeNext {
			current += string(command)
			escapeNext = false

			continue
		}

		// Detect escape character
		if command == '\\' {
			escapeNext = true

			continue
		}

		// Detect quote start
		if command == '"' || command == '\'' {
			state = stateQuotes
			quote = string(command)

			continue
		}

		// When in argument (not in quotes)
		if state == stateArg {
			if command == ' ' || command == '\t' {
				// Space ends current argument
				args = append(args, current)
				current = ""
				state = stateStart
			} else {
				current += string(command)
			}

			continue
		}

		// Detect start of new argument
		if command != ' ' && command != '\t' {
			state = stateArg
			current += string(command)
		}
	}

	// Error if we ended while still inside quotes
	if state == stateQuotes {
		return []string{}, fmt.Errorf("%w: %s", ErrUnclosedQuote, query)
	}

	// Add any remaining argument
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
