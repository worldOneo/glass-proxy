package tcpproxy

import "net"

// TCPHost contains a config and a status about this host
type TCPHost struct {
	Name   string
	Addr   string
	Status *HostStatus
}

// HostStatus contains *dynamic* information about a host e.g: Health
type HostStatus struct {
	Online bool
}

// NewTCPHost returns a new TCPHost
func NewTCPHost(name string, addr string) *TCPHost {
	return &TCPHost{
		Name: name,
		Addr: addr,
		Status: &HostStatus{
			Online: true,
		},
	}
}

// IsRunning tries to connect to that host and returns true or false if the host is running or not
func (T *TCPHost) IsRunning() bool {
	conn, err := net.Dial("tcp", T.Addr)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}
