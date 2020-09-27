package cmds

import (
	"fmt"

	"github.com/worldOneo/glass-proxy/src/config"
	"github.com/worldOneo/glass-proxy/src/tcpproxy"
)

// AddCmd is a command to add a server to the Proxy
type AddCmd struct {
	proxyService *tcpproxy.ProxyService
}

// NewAddCommand creates a new AddCmd
func NewAddCommand(proxyService *tcpproxy.ProxyService) *AddCmd {
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
	a.proxyService.Config.Hosts = append(a.proxyService.Config.Hosts, config.HostConfig{
		Name: name,
		Addr: addr,
	})
	a.proxyService.LoadHosts()
}
