package discordant

// Command is the Discord command.
type Command struct {
	Name        string   `json:"name"`
	Arg         string   `json:"arg"`
	Description string   `json:"description"`
	Help        string   `json:"help"`
	Access      []string `json:"access"`
	action      HandlerFunc
}

// CommandOption describes command option func.
type CommandOption func(*Command)

// MiddlewareAccess adds access levels.
func MiddlewareAccess(access ...string) CommandOption {
	return func(c *Command) {
		c.Access = append(c.Access, access...)
	}
}

// MiddlewareDescription adds description to command..
func MiddlewareDescription(description string) CommandOption {
	return func(c *Command) {
		c.Description = description
	}
}
