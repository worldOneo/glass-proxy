package tcpproxy

import (
	"net"
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
	Connections ProxyDict
}

// ProxyDict a map of all proxys
type ProxyDict map[*ReverseProxy]struct{}

// NewTCPHost returns a new TCPHost
func NewTCPHost(name string, addr string) *TCPHost {
	return &TCPHost{
		Name: name,
		Addr: addr,
		Status: &HostStatus{
			Online:      true,
			Connections: make(map[*ReverseProxy]struct{}),
		},
	}
}

// IsRunning tries to connect to that host and returns true or false if it is able to connect.
// This also updates the health of the host.
func (T *TCPHost) IsRunning() bool {
	T.HealthCheck()
	return T.Status.Online
}

// HealthCheck let this host perform a health check and updates it health information
func (T *TCPHost) HealthCheck() (bool, error) {
	conn, err := net.Dial("tcp", T.Addr)
	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()
	if err != nil {
		T.Status.Online = false
	} else {
		T.Status.Online = true
	}

	return T.Status.Online, err
}

// AddReverseProxy adds a new reverse proxy and starts it based on the connections given.
func (T *TCPHost) AddReverseProxy(conn net.Conn, serverConn net.Conn) {
	reverseProxy := NewReverseProxy(conn, serverConn)
	T.Status.Connections[reverseProxy] = struct{}{}
	defer func() {
		delete(T.Status.Connections, reverseProxy)
	}()
	reverseProxy.pipeBothAndClose()
}
