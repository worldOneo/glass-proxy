package cmds

// Command command#handle executes the command
type Command interface {
	handle([]string)
}
