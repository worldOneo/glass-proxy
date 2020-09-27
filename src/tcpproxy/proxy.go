package tcpproxy

import (
	"net"

	"github.com/worldOneo/glass-proxy/src/cmd"
	"github.com/worldOneo/glass-proxy/src/config"
	"github.com/worldOneo/glass-proxy/src/handler"
)

// ProxyService with everything we need
type ProxyService struct {
	Hosts          []*handler.TCPHost
	Config         *config.Config
	CommandHandler *cmd.CommandHandler
}

// ReverseProxy reverse tcp proxy
type ReverseProxy struct {
	biConn *handler.BiConn
}

// Pipe establisches a connection between both ends
func (r *ReverseProxy) Pipe() {
	go r.biConn.ConnectSend(1024)
	go r.biConn.ConnectRespond(1024)
}

// NewReverseProxy creates a new reverse tcp Proxy
func NewReverseProxy(conn1 net.Conn, conn2 net.Conn) *ReverseProxy {
	return &ReverseProxy{
		biConn: handler.NewBiConn(conn1, conn2),
	}
}

// LoadHosts populates ProxyService.Hosts from ProxyService.Config.Hosts
func (p *ProxyService) LoadHosts() {
	hosts := make([]*handler.TCPHost, 0)
	for _, host := range p.Config.Hosts {
		newHost := handler.NewTCPHost(host.Name, host.Addr)
		hosts = append(hosts, newHost)
	}
	p.Hosts = hosts
}
