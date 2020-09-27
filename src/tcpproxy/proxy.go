package tcpproxy

import (
	"net"

	"github.com/worldOneo/glass-proxy/src/handler"
)

// ReverseProxy reverse tcp proxy
type ReverseProxy struct {
	biConn *handler.BiConn
}

// Pipe establisches a connection between both ends
func (r *ReverseProxy) Pipe() {
	go r.biConn.ConnectSend(1024)
	go r.biConn.ConnectRespond(1024)
}

// NewReverseProxy creates a new reverse tcp Proxy
func NewReverseProxy(conn1 net.Conn, conn2 net.Conn) *ReverseProxy {
	return &ReverseProxy{
		biConn: handler.NewBiConn(conn1, conn2),
	}
}
