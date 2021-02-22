package udp

import (
	"log"
	"net"
	"sync"
	"time"

	"github.com/worldOneo/glass-proxy/proxy"
)

// Host type of proxy.Host with Connect for UDP
type Host interface {
	proxy.Host
	Connect([]byte, *net.UDPAddr, *net.UDPConn) error
	HealthCheck() (bool, error)
}

// host contains a config and a status about this host
type host struct {
	Name              string
	Addr              string
	Protocol          string
	Timeout           time.Duration
	UDPAddr           *net.UDPAddr
	Status            *HostStatus
	ClientServerCache *Cache
}

// HostStatus contains *dynamic* information about a host e.g: Health
type HostStatus struct {
	proxy.HostStatus
	sync.RWMutex
	Online      bool
	Connections int
}

// NewHost returns a new Host
func NewHost(name, addr, prot string, timeout int) Host {
	udpAddr, _ := net.ResolveUDPAddr("udp", addr)
	host := &host{
		Protocol:          prot,
		UDPAddr:           udpAddr,
		ClientServerCache: NewCache(time.Duration(timeout) * time.Second),
		Timeout:           time.Duration(timeout),
		Name:              name,
		Addr:              addr,
		Status: &HostStatus{
			Online: true,
		},
	}
	return host
}

// HealthCheck let this host perform a health check and updates it health information
func (U *host) HealthCheck() (bool, error) {
	return true, nil
}

func (U *host) GetAddr() string {
	return U.Addr
}

func (U *host) GetName() string {
	return U.Name
}

func (U *host) GetStatus() proxy.HostStatus {
	return U.Status
}

// IsOnline (not implemented) returns true
func (U *HostStatus) IsOnline() bool {
	return true
}

type udpData struct {
	buffer, oob []byte
}

func (U *host) Connect(buff []byte, clientaddr *net.UDPAddr, serviceconn *net.UDPConn) (err error) {
	var conn net.PacketConn

	testConn := U.ClientServerCache.Get(clientaddr)
	if testConn == nil {
		conn, err = net.ListenPacket(U.Protocol, ":0")
		if err != nil {
			log.Printf("Error dialing host: %v", err)
			return err
		}
		go U.Relay(conn, serviceconn, clientaddr)
		U.ClientServerCache.Put(clientaddr, conn)
	} else {
		conn = testConn.(*net.UDPConn)
	}

	if _, err = conn.WriteTo(buff, U.UDPAddr); err != nil {
		log.Printf("Unabel to forward packet to server (client->proxy-Xserver) %v", err)
		U.ClientServerCache.Remove(clientaddr)
	}
	return
}

// Relay forwards packets from downstream to upstream/toaddr
func (U *host) Relay(downstream net.PacketConn, upstream *net.UDPConn, toaddr *net.UDPAddr) {
	buffer := make([]byte, MUDS)
	U.Status.Lock()
	U.Status.Connections++
	U.Status.Unlock()

	defer func() {
		U.Status.Lock()
		U.Status.Connections--
		U.Status.Unlock()
		downstream.Close()
	}()

	log.Printf("Started relaying %s->%s", U.UDPAddr.String(), toaddr.String())

	for {
		downstream.SetReadDeadline(time.Now().Add(time.Millisecond * U.Timeout))
		lenb, _, err := downstream.ReadFrom(buffer)
		if err != nil {
			log.Printf("Unable to read datagram (server-Xproxy->client): %v", err)
			return
		}
		upstream.WriteToUDP(buffer[:lenb], toaddr)
		if err != nil {
			log.Printf("Unable to write datagram (server->proxy-Xclient): %v", err)
		}
	}
}

// GetConnectionCount returns the amount of connected hosts
func (U *HostStatus) GetConnectionCount() int {
	U.RLock()
	defer U.RUnlock()
	return U.Connections
}
