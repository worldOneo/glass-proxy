package tcp

import (
	"net"
	"sync"
	"time"

	"github.com/worldOneo/glass-proxy/proxy"
)

// Host type of proxy.Host with HealthCheck and AddReverseProxy
type Host interface {
	proxy.Host
	HealthCheck() (bool, error)
	AddReverseProxy(net.Conn, net.Conn)
}

// Host contains a config and a status about this host
type host struct {
	Name     string
	Addr     string
	Protocol string
	Status   *HostStatus
}

// HostStatus contains *dynamic* information about a host e.g: Health
type HostStatus struct {
	sync.RWMutex
	Online      bool
	Connections Dict
}

// Dict a map of all proxys
type Dict map[*ReverseProxy]struct{}

// NewHost returns a new Host
func NewHost(name string, addr string, protocol string) Host {
	host := &host{
		Name:     name,
		Addr:     addr,
		Protocol: protocol,
		Status: &HostStatus{
			Online:      true,
			Connections: make(map[*ReverseProxy]struct{}),
		},
	}
	return host
}

// IsRunning tries to connect to that host and returns true or false if it is able to connect.
// This also updates the health of the host.
func (T *host) IsRunning() bool {
	T.HealthCheck()
	return T.Status.Online
}

// HealthCheck let this host perform a health check and updates it health information
func (T *host) HealthCheck() (bool, error) {
	conn, err := net.DialTimeout(T.Protocol, T.Addr, 2*time.Second)

	defer func() {
		if conn != nil {
			conn.Close()
			conn = nil
		}
	}()

	T.Status.Lock()
	if err != nil {
		T.Status.Online = false
	} else {
		T.Status.Online = true
	}
	T.Status.Unlock()

	return T.Status.Online, err
}

// AddReverseProxy adds a new reverse proxy and starts it based on the connections given.
func (T *host) AddReverseProxy(conn net.Conn, serverConn net.Conn) {
	reverseProxy := NewReverseProxy(conn, serverConn)

	T.Status.Lock()
	T.Status.Connections[reverseProxy] = struct{}{}
	T.Status.Unlock()
	defer func() {
		T.Status.Lock()
		delete(T.Status.Connections, reverseProxy)
		T.Status.Unlock()
	}()
	reverseProxy.pipeBothAndClose()
}

// GetConnectionCount returns the amount of connections held by this Host
func (T *HostStatus) GetConnectionCount() int {
	T.RLock()
	defer T.RUnlock()
	return len(T.Connections)
}

// IsOnline returns if the host was able to connect to its address
func (T *host) IsOnline() bool {
	return T.Status.IsOnline()
}

// IsOnline returns if the host is online
func (T *HostStatus) IsOnline() bool {
	T.RLock()
	defer T.RUnlock()
	return T.Online
}

// GetName returns the name of the host
func (T *host) GetName() string {
	return T.Name
}

// GetAddr returns the remote addr of the host
func (T *host) GetAddr() string {
	return T.Addr
}

func (T *host) GetStatus() proxy.HostStatus {
	return T.Status
}
