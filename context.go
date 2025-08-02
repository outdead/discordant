package discordant

import (
	"bufio"
	"bytes"
	ctx "context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/bwmarrin/discordgo"
)

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
	QuerySlice() ([]string, error)
	QueryAttachmentBodyFirst() (string, error)
	Send(msg string, params ...string) error
	Success() error
	Fail() error
	JSON(rawmsg any, params ...string) error
	JSONPretty(rawmsg any, params ...string) error
	Embed(msg *discordgo.MessageEmbed) error
	Embeds(msgs []discordgo.MessageEmbed) error
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

// QuerySlice parses the command query string and returns a slice of arguments.
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
func (c *context) QuerySlice() ([]string, error) { //nolint: cyclop, funlen // indivisible
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

// QueryAttachmentBodyFirst retrieves the content of the first attachment from a message.
// It performs the following operations:
//  1. Checks if there are any attachments - returns empty string if none exist
//  2. Fetches the content from the URL of the first attachment
//  3. Returns the content as a string
//
// This is typically used to process message attachments where only the first attachment's
// content is needed (e.g., processing a single file upload).
//
// Returns:
//   - string: The content of the first attachment as a string
//   - error: Any error that occurred during the HTTP request or content reading
//     (network errors, invalid URL, read errors, etc.)
//
// The function automatically closes the response body after reading.
func (c *context) QueryAttachmentBodyFirst() (string, error) {
	if len(c.Request().Message.Attachments) == 0 {
		return "", ErrNoAttachment
	}

	uri := c.Request().Message.Attachments[0].URL

	req, err := http.NewRequestWithContext(ctx.Background(), http.MethodGet, uri, http.NoBody)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	buf := &bytes.Buffer{}
	if _, err := io.Copy(buf, resp.Body); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// Send sends message to discord channel.
// That is, what to do with messages that more than 2000 characters.
// 1. Send as file.
// 2. Send as multiple messages.
func (c *context) Send(msg string, params ...string) error {
	// Send normal message.
	if len([]rune(msg)) <= DiscordMaxMessageLenValidate {
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

// Success sends a success response message.
func (c *context) Success() error {
	return c.Send(ResponseMessageSuccess)
}

// Fail sends a fail response message.
func (c *context) Fail() error {
	return c.Send(ResponseMessageFail)
}

// JSON sends a JSON response with the given message/object.
// It converts the input to JSON format and sends it as a string response.
// The input can be a string, error, or any type that can be marshaled to JSON.
// Optional parameters can be provided for additional response configuration.
func (c *context) JSON(rawmsg any, params ...string) error {
	return c.json(rawmsg, false, params...)
}

// JSONPretty sends a JSON response with pretty-printed formatting (indented).
// Similar to JSON() but outputs human-readable formatted JSON.
// The input can be a string, error, or any type that can be marshaled to JSON.
// Optional parameters can be provided for additional response configuration.
func (c *context) JSONPretty(rawmsg any, params ...string) error {
	return c.json(rawmsg, true, params...)
}

// Embed sends a message with embedded data.
func (c *context) Embed(msg *discordgo.MessageEmbed) error {
	if _, err := c.discordant.session.ChannelMessageSendEmbed(c.ChannelID(), msg); err != nil {
		return err
	}

	return nil
}

// Embeds sends messages with embedded data.
func (c *context) Embeds(msgs []discordgo.MessageEmbed) error {
	if len(msgs) == 0 {
		return ErrEmptyResponseMessage
	}

	for i := range msgs {
		msg := msgs[i]

		if err := c.Embed(&msg); err != nil {
			return err
		}
	}

	return nil
}

// json is the internal implementation for JSON response handling.
// It handles the conversion of different input types to JSON format and sends the response.
// The pretty parameter controls whether the JSON output is formatted with indentation.
func (c *context) json(rawmsg any, pretty bool, params ...string) error {
	var msg string

	switch val := rawmsg.(type) {
	case string:
		msg = fmt.Sprintf(ResponseMessageFormatJSON, val)
	case error:
		msg = fmt.Sprintf(ResponseMessageFormatJSON, val.Error())
	default:
		var (
			rawjson []byte
			err     error
		)

		if pretty {
			rawjson, err = json.MarshalIndent(val, "", "  ")
		} else {
			rawjson, err = json.Marshal(val)
		}

		if err != nil {
			return fmt.Errorf("%w: %w", ErrInvalidResponseMessageType, err)
		}

		msg = fmt.Sprintf(ResponseMessageFormatJSON, string(rawjson))
	}

	return c.Send(msg, params...)
}
