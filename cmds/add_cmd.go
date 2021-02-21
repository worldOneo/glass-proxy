package cmds

import (
	"fmt"

	"github.com/worldOneo/glass-proxy/config"
	"github.com/worldOneo/glass-proxy/proxy"
)

// AddCmd is a command to add a server to the Proxy
type AddCmd struct {
	proxyService proxy.Service
}

// NewAddCommand creates a new AddCmd
func NewAddCommand(proxyService proxy.Service) *AddCmd {
	return &AddCmd{
		proxyService: proxyService,
	}
}

// Handle handles the commands and adds it to the ProxyService it was initialized with
func (a *AddCmd) Handle(args []string) {
	if len(args) < 2 {
		fmt.Println("\"add\" needs 2 args the name of the server and the address")
		return
	}

	name := args[0]
	addr := args[1]
	a.proxyService.AddHost(config.HostConfig{
		Name: name,
		Addr: addr,
	})
}
