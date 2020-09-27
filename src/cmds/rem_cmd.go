package cmds

import (
	"fmt"

	"github.com/worldOneo/glass-proxy/src/config"
	"github.com/worldOneo/glass-proxy/src/tcpproxy"
)

// RemCmd is a command to remove a server from the Proxy
type RemCmd struct {
	proxyService *tcpproxy.ProxyService
}

// NewRemCommand creates a new AddCmd
func NewRemCommand(proxyService *tcpproxy.ProxyService) *RemCmd {
	return &RemCmd{
		proxyService: proxyService,
	}
}

// Handle handles the commands and adds it to the ProxyService it was initialized with
func (r *RemCmd) Handle(args []string) {
	if len(args) < 1 {
		fmt.Println("\"rem\" needs 1 arg, the name of the server")
		return
	}

	name := args[0]
	hosts := make([]config.HostConfig, 0)
	for _, host := range r.proxyService.Config.Hosts {
		if host.Name != name {
			hosts = append(hosts, host)
		}
	}
	r.proxyService.Config.Hosts = hosts
	r.proxyService.LoadHosts()
}
