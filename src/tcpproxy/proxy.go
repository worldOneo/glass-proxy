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
	go r.pipeBothAndClose()
}

func (r *ReverseProxy) pipeBothAndClose() {
	go r.biConn.ConnectSend()
	r.biConn.ConnectRespond()
	r.biConn.Conn1.Close()
	r.biConn.Conn2.Close()
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

// AddHost adds a host and adds it to the config
func (p *ProxyService) AddHost(host config.HostConfig) {
	p.Config.Hosts = append(p.Config.Hosts, host)
	p.LoadHosts()
}

// RemHost removes a host
func (p *ProxyService) RemHost(name string) {
	hosts := make([]config.HostConfig, 0)
	for _, host := range p.Config.Hosts {
		if host.Name != name {
			hosts = append(hosts, host)
		}
	}
	p.Config.Hosts = hosts
	p.LoadHosts()
}
