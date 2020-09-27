package cmd

import (
	"bufio"
	"os"
	"strings"
)

// CommandHandler holds commands which are registered
type CommandHandler struct {
	commandMap map[string]func([]string)
}

const helpText = "====COMMANDS====\nstop stops the proxy\nadd <NAME> <ADDR> Add a server\nrem <NAME> Remove a server\nlist show all servers"

// NewCommandHandler creates a new CommandHandler
func NewCommandHandler() *CommandHandler {
	return &CommandHandler{
		commandMap: make(map[string]func([]string)),
	}
}

// Handle gets the command from the string and executes it
func (c *CommandHandler) Handle(str string) {
	str = strings.TrimLeft(str, "\n\r \t")
	str = strings.TrimRight(str, "\n\r \t")
	if len(str) == 0 {
		return
	}
	args := strings.Split(str, " ")
	if len(args) == 0 {
		return
	}
	cmd, exist := c.commandMap[strings.ToLower(args[0])]
	if !exist {
		println(helpText)
		return
	}
	cmd(args[1:])
}

// Register registers the new command as cmd
func (c *CommandHandler) Register(cmd string, f func([]string)) {
	c.commandMap[cmd] = f
}

// Listen listen to the commandline input
func (c *CommandHandler) Listen() {
	for {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			text := scanner.Text()
			if strings.ToLower(text) == "stop" {
				return
			}
			c.Handle(text)
		}
	}
}
