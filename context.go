package discordant

import (
	"bufio"
	"bytes"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// Context is an interface represents the context of the current Discord command.
type Context interface {
	Command() *Command
	Discordant() *Discordant
	Request() *discordgo.MessageCreate
	ChannelID() string
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
	if len([]rune(msg)) <= 1990 {
		_, err := c.discordant.session.ChannelMessageSend(c.request.ChannelID, msg)

		return err
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
		return err
	}

	ms := &discordgo.MessageSend{Files: []*discordgo.File{
		{Name: fileName, Reader: bufio.NewReader(&buf)},
	}}

	_, err := c.discordant.session.ChannelMessageSendComplex(c.request.ChannelID, ms)

	return err
}
