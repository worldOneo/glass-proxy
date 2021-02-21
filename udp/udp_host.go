package udp

import (
	"log"
	"net"
	"sync"
	"time"

	"github.com/worldOneo/glass-proxy/proxy"
)

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
	UDPAddr           *net.UDPAddr
	Status            *HostStatus
	ClientServerCache *Cache
}

// HostStatus contains *dynamic* information about a host e.g: Health
type HostStatus struct {
	proxy.HostStatus
	sync.RWMutex
	Online      bool
	Connections map[*net.IPAddr]*net.Conn
}

// NewHost returns a new Host
func NewHost(name, addr, prot string) Host {
	udpAddr, _ := net.ResolveUDPAddr("udp", addr)
	host := &host{
		Protocol:          prot,
		UDPAddr:           udpAddr,
		ClientServerCache: NewCache(5 * time.Second),
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

func (U *HostStatus) IsOnline() bool {
	U.RLock()
	defer U.RUnlock()
	return U.Online
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
		go Relay(conn, serviceconn, clientaddr, U.UDPAddr)
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

func Relay(downstream net.PacketConn, upstream *net.UDPConn, toaddr, fromaddr *net.UDPAddr) {
	buffer := make([]byte, MUDS)

	defer func() {
		downstream.Close()
	}()

	log.Printf("Started relaying %s->%s", fromaddr.String(), toaddr.String())

	for {
		downstream.SetReadDeadline(time.Now().Add(time.Second * 3))
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
