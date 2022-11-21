package discordant

import (
	"errors"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/outdead/discordant/internal"
)

// Channel types.
const (
	ChannelAdmin   = "admin"
	ChannelGeneral = "general"
)

// Defaults.
const (
	DefaultCommandPrefix    = "!"
	DefaultCommandDelimiter = " "
)

// Response massage layouts.
const (
	ResponseMessageFail       = "```fail```"
	ResponseMessageFormatJSON = "```json\n%s\n```"
)

// DiscordMaxMessageLen max discord message length.
const DiscordMaxMessageLen = 2000

var (
	// ErrEmptyToken is returned when discord bot token is empty with enabled hook.
	ErrEmptyToken = errors.New("discord bot token is empty")

	// ErrEmptyPrefix is returned when discord bot prefix is empty.
	ErrEmptyPrefix = errors.New("discord bot prefix is empty")

	// ErrEmptyChannelID is returned when discord channel id is empty with enabled hook.
	ErrEmptyChannelID = errors.New("discord channel id is empty")

	// ErrMessageTooLong is returned when message that has been sent to discord longer
	// than 2000 characters.
	ErrMessageTooLong = errors.New("discord message too long")

	// ErrCommandNotFound is returned if an unknown command was received.
	ErrCommandNotFound = errors.New("command not found")
)

// HandlerFunc defines a function to serve HTTP requests.
type HandlerFunc func(Context) error

// Discordant represents a connection to the Discord API.
type Discordant struct {
	config              *Config
	id                  string
	session             *internal.Session
	logger              Logger
	commands            map[string]Command
	commandsAccessOrder []string
}

// New creates a new Discord session and will automate some startup
// tasks if given enough information to do so. Currently, you can pass zero
// arguments, and it will return an empty Discord session.
func New(cfg *Config, options ...Option) (*Discordant, error) {
	d := Discordant{
		config:   cfg,
		logger:   NewDefaultLog(),
		commands: make(map[string]Command),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	for _, option := range options {
		option(&d)
	}

	if d.session == nil && cfg.Token == "" {
		return nil, ErrEmptyToken
	}

	if d.session == nil {
		var err error
		if d.session, err = internal.NewSession(cfg.Token); err != nil {
			return nil, fmt.Errorf("discordant: %w", err)
		}
	}

	user, err := d.session.User("@me")
	if err != nil {
		_ = d.Close()

		return nil, fmt.Errorf("discordant: retrieve bot account: %w", err)
	}

	d.id = user.ID

	if len(d.config.AccessOrder) == 0 {
		d.commandsAccessOrder = []string{ChannelGeneral, ChannelAdmin}
	} else {
		copy(d.commandsAccessOrder, d.config.AccessOrder)
	}

	return &d, nil
}

// Close closes discord connection.
func (d *Discordant) Close() error {
	if d.session != nil {
		if err := d.session.Close(); err != nil {
			return fmt.Errorf("discordant: close connection: %w", err)
		}
	}

	return nil
}

// Run runs discord bot handlers.
func (d *Discordant) Run() {
	for name, command := range d.commands {
		command.Name = name
		d.commands[name] = command
	}

	d.AddHandler(d.commandHandler)
}

// ID returns stored bot id.
func (d *Discordant) ID() string {
	return d.id
}

// Session returns discord Session.
func (d *Discordant) Session() *discordgo.Session {
	return d.session.Session
}

// Commands returns commands list.
func (d *Discordant) Commands() map[string]Command {
	return d.commands
}

// AddHandler allows you to add an event handler that will be fired anytime
// the Discord WSAPI event that matches the function fires.
func (d *Discordant) AddHandler(handler interface{}) func() {
	return d.session.AddHandler(handler)
}

// NewContext creates new Context.
func (d *Discordant) NewContext(message *discordgo.MessageCreate, command *Command) Context {
	return &context{
		command:    command,
		request:    message,
		discordant: d,
	}
}

// ADMIN adds route handler to admin channel.
func (d *Discordant) ADMIN(name string, handler HandlerFunc, options ...CommandOption) {
	options = append(options, MiddlewareAccess(ChannelAdmin))
	d.Add(name, handler, options...)
}

// GENERAL adds route handler to general channel.
func (d *Discordant) GENERAL(name string, handler HandlerFunc, options ...CommandOption) {
	options = append(options, MiddlewareAccess(ChannelGeneral))
	d.Add(name, handler, options...)
}

// ALL adds route handler to any channel.
func (d *Discordant) ALL(name string, handler HandlerFunc, options ...CommandOption) {
	options = append(options, MiddlewareAccess(ChannelGeneral, ChannelAdmin))
	d.Add(name, handler, options...)
}

// Add adds route handler.
func (d *Discordant) Add(name string, handler HandlerFunc, options ...CommandOption) {
	command := Command{
		action: handler,
	}

	for _, option := range options {
		option(&command)
	}

	d.fixCommandAccess(&command)

	d.commands[name] = command
}

// GetCommand returns command by received message.
func (d *Discordant) GetCommand(message string) (*Command, error) {
	// Command without args.
	if command, ok := d.commands[message]; ok {
		return &command, nil
	}

	// Find commands with args.
	for name, command := range d.commands {
		if strings.Index(message, name+DefaultCommandDelimiter) == 0 {
			// Workaround for intersecting commands such as `rules` and `rules set` or `help` and `help set`.
			// TODO: Come up with a better solution.
			if strings.Index(message, name+DefaultCommandDelimiter+"set ") == 0 {
				continue
			}

			command.Arg = strings.Replace(message, name, "", 1)
			command.Arg = strings.TrimSpace(command.Arg)

			return &command, nil
		}
	}

	return nil, ErrCommandNotFound
}

// CheckAccess returns true if access is allowed.
func (d *Discordant) CheckAccess(id string, channels ...string) bool {
	if len(channels) == 0 {
		return true
	}

	for _, channel := range channels {
		if id == d.config.Channels[channel] {
			return true
		}
	}

	return false
}

func (d *Discordant) commandHandler(session *discordgo.Session, message *discordgo.MessageCreate) {
	// Do nothing because the bot is talking.
	if message.Author.Bot || message.Author.ID == d.id {
		return
	}

	// Not bot command. Do nothing.
	if strings.Index(message.Content, d.config.Prefix) != 0 {
		return
	}

	if d.config.Safemode {
		// Unknown channel. Do nothing.
		if !d.CheckAccess(message.ChannelID, ChannelGeneral, ChannelAdmin) {
			d.logger.Debugf("unknown channel %s", message.ChannelID)

			return
		}
	}

	// Remove prefix from discord message.
	content := strings.TrimPrefix(message.Content, d.config.Prefix)

	command, err := d.GetCommand(content)
	if err != nil {
		d.logger.Debug(err)

		return
	}

	if ok := d.CheckAccess(message.ChannelID, command.Access...); !ok {
		d.logger.Debugf("access to command \"%s\" denied", command.Name)

		return
	}

	ctx := d.NewContext(message, command)

	if err := command.action(ctx); err != nil {
		d.logger.Error(err)

		if err := ctx.Send(ResponseMessageFail); err != nil {
			d.logger.Errorf("send fail response error: %s", err)
		}
	}
}

func (d *Discordant) fixCommandAccess(command *Command) {
	buf := make(map[string]struct{}, len(d.commandsAccessOrder))

	access := make([]string, 0, len(d.commandsAccessOrder))

	for _, channel := range command.Access {
		if _, ok := buf[channel]; !ok {
			buf[channel] = struct{}{}
		}
	}

	for _, channel := range d.commandsAccessOrder {
		if _, ok := buf[channel]; ok {
			access = append(access, channel)
		}
	}

	command.Access = access
}
