package tcpproxy

import (
	"errors"
	"math/rand"
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

// GetHost gets a random running host or nil if no host is available
func (p *ProxyService) GetHost() *handler.TCPHost {
	l := len(p.Hosts)
	if l == 0 {
		return nil
	}
	i := rand.Intn(l)
	h := p.Hosts[i]
	if h.Status.Online {
		return h
	}
	mx := i + l
	for j := i; j < mx; j++ {
		h = p.Hosts[j%l]
		if h.Status.Online {
			return h
		}
	}
	return nil
}

// Dial dials a connection or returns error.
// Uses the default inreface or itterates over every given and tries to dial over it if an interface is given
func (p *ProxyService) Dial(protocol string, addr string) (net.Conn, error) {
	if p.Config.Interfaces == nil || len(p.Config.Interfaces) == 0 {
		return net.Dial(protocol, addr)
	}

	err := errors.New("Couldn't dial a connection over any of the given interfaces")
	for _, i := range p.Config.Interfaces {
		ief, err := net.InterfaceByName(i)
		if err != nil {
			continue
		}
		addrs, err := ief.Addrs()
		if err != nil {
			continue
		}
		for _, pladdr := range addrs {
			var targetIP string
			var resError error
			ip, _, err := net.ParseCIDR(pladdr.String())
			if err != nil {
				return nil, err
			}
			if ip.IsUnspecified() {
				continue
			}
			if ip.To4().Equal(ip) {
				targetIP = ip.String()
			} else {
				targetIP = "[" + ip.String() + "]"
			}
			d := net.Dialer{}
			d.LocalAddr, resError = net.ResolveTCPAddr(protocol, targetIP+":0")
			if resError != nil {
				continue
			}
			conn, err := d.Dial(protocol, addr)
			if err != nil {
				continue
			}
			return conn, err
		}
	}
	return nil, err
}
