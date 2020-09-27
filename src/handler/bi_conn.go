package handler

import (
	"io"
	"net"
)

// BiConn contains both connections
type BiConn struct {
	Conn1 net.Conn
	Conn2 net.Conn
}

// ConnectSend Starts the Sending from conn1 to conn2
func (b *BiConn) ConnectSend(buffSize int) error {
	return pipe(b.Conn1, b.Conn2, buffSize)
}

func pipe(conn1 net.Conn, conn2 net.Conn, buffSize int) error {
	_, err := io.Copy(conn1, conn2)
	return err
}

// ConnectRespond Starts the sending from conn2 to conn1
func (b *BiConn) ConnectRespond(buffSize int) error {
	return pipe(b.Conn2, b.Conn1, buffSize)
}

// NewBiConn creates a new BiConnection
func NewBiConn(conn1 net.Conn, conn2 net.Conn) *BiConn {
	return &BiConn{
		Conn1: conn1,
		Conn2: conn2,
	}
}
