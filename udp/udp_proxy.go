package udp

import (
	"errors"
	"log"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/worldOneo/glass-proxy/cmd"
	"github.com/worldOneo/glass-proxy/config"
	"github.com/worldOneo/glass-proxy/proxy"
)

// MUDS defines the Maximum UDP Datagram Size
const MUDS = 1200

var ebuff = make([]byte, MUDS)

// Service with everything we need
type Service struct {
	sync.RWMutex
	Cache          *Cache
	Hosts          []Host
	HostsLock      *sync.RWMutex
	Config         *config.Config
	CommandHandler *cmd.CommandHandler
}

// NewService creates a new Proxy Service and starts the cleaner
func NewService(cnf *config.Config) *Service {
	proxy := &Service{
		Cache:          NewCache(3 * time.Second),
		Config:         cnf,
		CommandHandler: cmd.NewCommandHandler(),
		HostsLock:      &sync.RWMutex{},
	}
	proxy.LoadHosts()

	return proxy
}

// LoadHosts populates Service.Hosts from Service.Config.Hosts
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

func (p *Service) Handle(clientaddr *net.UDPAddr, datagram []byte, serviceconn *net.UDPConn) error {
	host := p.Cache.Get(clientaddr)
	if host == nil {
		host = p.GetHost()
		if host == nil {
			return errors.New("no healthy host available")
		}
		p.Cache.Put(clientaddr, host)
	}
	newDG := make([]byte, len(datagram))
	copy(newDG, datagram)
	go host.(Host).Connect(newDG, clientaddr, serviceconn)
	return nil
}

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

func (p *Service) Run() error {
	laddr, err := net.ResolveUDPAddr("udp", p.Config.Addr)
	if err != nil {
		log.Fatalf("Unabele to resolve \"%s\"", laddr)
		return err
	}

	go p.HealthCheck()

	serviceconn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		log.Fatalf("Unabele to listen on \"%s\"", laddr)
		return err
	}
	log.Print("Started")

	datagram := make([]byte, MUDS)
	for {
		ldat, clientaddr, err := serviceconn.ReadFromUDP(datagram)
		if err != nil {
			log.Printf("Unable to read datagram (client-Xproxy->server) %v", err)
			continue
		}
		p.Handle(clientaddr, datagram[:ldat], serviceconn)
		copy(datagram, ebuff)
	}
}

// GetHost gets a random running host or nil if no host is available
func (p *Service) GetHost() Host {
	p.HostsLock.RLock()
	defer p.HostsLock.RUnlock()
	l := len(p.Hosts)
	if l == 0 {
		return nil
	}
	i := rand.Intn(l)
	return p.Hosts[i]
}

func (p *Service) AddHost(hostconfig config.HostConfig) {
	p.HostsLock.Lock()
	defer p.HostsLock.Unlock()
	host := NewHost(hostconfig.Name, hostconfig.Addr, p.Config.Protocol)
	p.Hosts = append(p.Hosts, host)
}

func (p *Service) RemHost(name string) {
	p.HostsLock.Lock()
	hosts := make([]config.HostConfig, 0)
	for _, host := range p.Config.Hosts {
		if host.Name != name {
			hosts = append(hosts, host)
		}
	}
	p.Config.Hosts = hosts
	p.HostsLock.Unlock()
	p.LoadHosts()
}

func (p *Service) GetConfig() *config.Config {
	return p.Config
}

func (p *Service) ListHosts() []proxy.Host {
	p.HostsLock.RLock()
	defer p.HostsLock.RUnlock()
	castedHosts := make([]proxy.Host, 0)
	for _, r := range p.Hosts {
		castedHosts = append(castedHosts, r)
	}
	return castedHosts
}
