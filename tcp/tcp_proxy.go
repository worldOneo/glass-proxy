package tcp

import (
	"errors"
	"log"
	"math"
	"net"
	"sync"
	"time"

	"github.com/worldOneo/glass-proxy/cmd"
	"github.com/worldOneo/glass-proxy/config"
	"github.com/worldOneo/glass-proxy/handler"
	"github.com/worldOneo/glass-proxy/proxy"
)

// Service with everything we need
type Service struct {
	proxy.Service
	Hosts          []Host
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
func NewProxyService(cnf *config.Config) *Service {
	proxy := &Service{
		Config:         cnf,
		CommandHandler: cmd.NewCommandHandler(),
		HostsLock:      &sync.RWMutex{},
	}
	proxy.LoadHosts()

	return proxy
}

// LoadHosts populates ProxyService.Hosts from ProxyService.Config.Hosts
func (p *Service) LoadHosts() {
	p.HostsLock.Lock()
	defer p.HostsLock.Unlock()
	hosts := make([]Host, 0)
	for _, host := range p.Config.Hosts {
		newHost := NewHost(host.Name, host.Addr, p.Config.Protocol)
		hosts = append(hosts, newHost)
	}
	p.Hosts = hosts
}

// AddHost adds a host and adds it to the config
func (p *Service) AddHost(host config.HostConfig) {
	p.HostsLock.Lock()
	defer p.HostsLock.Unlock()
	p.Config.Hosts = append(p.Config.Hosts, host)
	p.LoadHosts()
}

// RemHost removes a host
func (p *Service) RemHost(name string) {
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
func (p *Service) GetHost() Host {
	p.HostsLock.RLock()
	defer p.HostsLock.RUnlock()

	index := -1
	min := math.MaxInt32
	c := 0
	for i, h := range p.Hosts {
		c = h.GetStatus().GetConnectionCount()
		if c < min && h.GetStatus().IsOnline() {
			min = c
			index = i
		}
	}
	if index == -1 {
		return nil
	}
	return p.Hosts[index]
}

// DialToHost dials a connection or returns error.
// Uses the default inreface or itterates over every given and tries to dial over it if an interface is given
func (p *Service) DialToHost(protocol string, client net.Conn) (Host, net.Conn, error) {
	host := p.GetHost()
	if host == nil {
		return nil, nil, errors.New("No Healthy host available")
	}
	addr := host.GetAddr()
	if p.Config.Interfaces == nil || len(p.Config.Interfaces) == 0 {
		conn, err := net.Dial(protocol, addr)
		return host, conn, err
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
			return host, conn, err
		}
	}
	return nil, nil, err
}

// HealthCheck checks the health of every given server and updates their status
func (p *Service) HealthCheck() {
	for {
		p.HostsLock.RLock()
		for _, h := range p.Hosts {
			h.HealthCheck()
		}
		p.HostsLock.RUnlock()
		time.Sleep(time.Second * time.Duration(p.Config.HealthCheckTime))
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

func (p *Service) Handle(conn net.Conn) {
	host, remote, err := p.DialToHost(p.Config.Protocol, conn)

	if err != nil {
		if host == nil {
			log.Printf("Couldn't connect to any host \"%v\"", err)
			return
		}
		log.Printf("Couldn't connect to %s (%s) \"%v\"", host.GetName(), host.GetAddr(), err)
		return
	}

	defer func() {
		if p.Config.LogConfig.LogDisconnect {
			log.Printf("%s Disconnected", conn.RemoteAddr())
		}
		conn.Close()
	}()

	if err != nil {
		log.Printf("Couldn't connect to %s (%s) \"%v\"", host.GetName(), host.GetAddr(), err)
		return
	}

	if p.Config.LogConfig.LogConnections {
		log.Printf("%s Connected to %s (%s) over %s", conn.RemoteAddr(), host.GetName(), host.GetAddr(), remote.LocalAddr())
	}

	host.AddReverseProxy(conn, remote)
}

// ListHosts gets all hosts useable for this service
func (p *Service) ListHosts() []proxy.Host {
	p.HostsLock.RLock()
	defer p.HostsLock.RUnlock()
	castedHosts := make([]proxy.Host, 0)
	for _, r := range p.Hosts {
		castedHosts = append(castedHosts, r)
	}
	return castedHosts
}

func (p *Service) GetConfig() *config.Config {
	return p.Config
}

func (p *Service) Run() {
	go p.HealthCheck()
	ln, err := net.Listen(p.Config.Protocol, p.Config.Addr)
	if err != nil {
		log.Fatalf("Couldn't start the server: %v", err)
	}
	log.Printf("Listening on %s", p.Config.Addr)
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go p.Handle(conn)
	}
}
