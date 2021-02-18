package tcpproxy

import (
	"errors"
	"math/rand"
	"net"
	"sync"

	"github.com/worldOneo/glass-proxy/cmd"
	"github.com/worldOneo/glass-proxy/config"
	"github.com/worldOneo/glass-proxy/handler"
)

// ProxyService with everything we need
type ProxyService struct {
	Hosts          []*TCPHost
	HostsLock      *sync.RWMutex
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

// NewProxyService creates a new Proxy Service and starts the cleaner
func NewProxyService(cnf *config.Config) *ProxyService {
	proxy := &ProxyService{
		Config:         cnf,
		CommandHandler: cmd.NewCommandHandler(),
		HostsLock:      &sync.RWMutex{},
	}
	proxy.LoadHosts()

	return proxy
}

// LoadHosts populates ProxyService.Hosts from ProxyService.Config.Hosts
func (p *ProxyService) LoadHosts() {
	p.HostsLock.Lock()
	defer p.HostsLock.Unlock()
	hosts := make([]*TCPHost, 0)
	for _, host := range p.Config.Hosts {
		newHost := NewTCPHost(host.Name, host.Addr)
		hosts = append(hosts, newHost)
	}
	p.Hosts = hosts
}

// AddHost adds a host and adds it to the config
func (p *ProxyService) AddHost(host config.HostConfig) {
	p.HostsLock.Lock()
	defer p.HostsLock.Unlock()
	p.Config.Hosts = append(p.Config.Hosts, host)
	p.LoadHosts()
}

// RemHost removes a host
func (p *ProxyService) RemHost(name string) {
	p.HostsLock.Lock()
	defer p.HostsLock.Unlock()
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
func (p *ProxyService) GetHost() *TCPHost {
	p.HostsLock.RLock()
	defer p.HostsLock.RUnlock()
	l := len(p.Hosts)
	if l == 0 {
		return nil
	}
	i := rand.Intn(l)
	h := p.Hosts[i]
	if h.IsOnline() {
		return h
	}
	mx := i + l
	for j := i; j < mx; j++ {
		h = p.Hosts[j%l]
		if h.IsOnline() {
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
			d, dialError := createDialer(protocol, pladdr.String())
			if dialError != nil {
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

// HealthCheck checks the health of every given server and updates their status
func (p *ProxyService) HealthCheck() {
	p.HostsLock.RLock()
	defer p.HostsLock.RUnlock()
	for _, h := range p.Hosts {
		h.HealthCheck()
	}
}

func createDialer(protocol, addr string) (*net.Dialer, error) {
	var resError error
	var targetIP string
	ip, _, err := net.ParseCIDR(addr)
	if err != nil {
		return nil, err
	}
	if ip.To4().Equal(ip) {
		targetIP = ip.String()
	} else {
		targetIP = "[" + ip.String() + "]"
	}
	d := net.Dialer{}
	d.LocalAddr, resError = net.ResolveTCPAddr(protocol, targetIP+":0")
	return &d, resError
}
