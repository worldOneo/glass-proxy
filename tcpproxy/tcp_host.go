package tcpproxy

import (
	"net"
	"sync"
)

// TCPHost contains a config and a status about this host
type TCPHost struct {
	Name   string
	Addr   string
	Status *HostStatus
}

// HostStatus contains *dynamic* information about a host e.g: Health
type HostStatus struct {
	Online      bool
	StatusLock  *sync.RWMutex
	Connections ProxyDict
}

// ProxyDict a map of all proxys
type ProxyDict map[*ReverseProxy]struct{}

// NewTCPHost returns a new TCPHost
func NewTCPHost(name string, addr string) *TCPHost {
	host := &TCPHost{
		Name: name,
		Addr: addr,
		Status: &HostStatus{
			Online:      true,
			Connections: make(map[*ReverseProxy]struct{}),
			StatusLock:  &sync.RWMutex{},
		},
	}
	return host
}

// IsRunning tries to connect to that host and returns true or false if it is able to connect.
// This also updates the health of the host.
func (T *TCPHost) IsRunning() bool {
	T.HealthCheck()
	return T.Status.Online
}

// HealthCheck let this host perform a health check and updates it health information
// memleak
func (T *TCPHost) HealthCheck() (bool, error) {
	conn, err := net.Dial("tcp", T.Addr)

	defer func() {
		if conn != nil {
			conn.Close()
			conn = nil
		}
		err = nil
	}()

	T.Status.StatusLock.Lock()
	if err != nil {
		T.Status.Online = false
	} else {
		T.Status.Online = true
	}
	T.Status.StatusLock.Unlock()

	return T.Status.Online, err
}

// AddReverseProxy adds a new reverse proxy and starts it based on the connections given.
func (T *TCPHost) AddReverseProxy(conn net.Conn, serverConn net.Conn) {
	reverseProxy := NewReverseProxy(conn, serverConn)
	T.Status.StatusLock.Lock()
	T.Status.Connections[reverseProxy] = struct{}{}
	T.Status.StatusLock.Unlock()
	defer func() {
		T.Status.StatusLock.Lock()
		delete(T.Status.Connections, reverseProxy)
		T.Status.StatusLock.Unlock()
	}()
	reverseProxy.pipeBothAndClose()
}

// GetConnectionCount returns the amount of connections held by this Host
func (T *TCPHost) GetConnectionCount() int {
	T.Status.StatusLock.RLock()
	defer T.Status.StatusLock.RUnlock()
	return len(T.Status.Connections)
}

// IsOnline returns if the host was able to connect to its address
func (T *TCPHost) IsOnline() bool {
	T.Status.StatusLock.RLock()
	defer T.Status.StatusLock.RUnlock()
	return T.Status.Online
}
