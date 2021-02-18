package main

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/worldOneo/glass-proxy/config"
	"github.com/worldOneo/glass-proxy/tcpproxy"
)

var tServers = []int{25555, 25556, 25557, 25558, 25559}

func TestProxy(t *testing.T) {
	hosts := make([]config.HostConfig, 0)

	for _, s := range tServers {
		t.Logf("Starting %d", s)
		addr := fmt.Sprintf("127.0.0.1:%d", s)
		go startEchoServer(addr, t)
		hosts = append(hosts, config.HostConfig{
			Name: addr,
			Addr: addr,
		})
	}
	rand.Seed(time.Now().UnixNano())

	proxyService := tcpproxy.NewProxyService(&config.Config{
		Addr:            "127.0.0.1:25560",
		Hosts:           hosts,
		HealthCheckTime: 1,
		LogConfig: config.LogConfig{
			LogConnections: false,
			LogDisconnect:  false,
		},
	})

	go start(proxyService)
	go healthCheck(proxyService)

	time.Sleep(time.Duration(5 * time.Second)) //await health check
	wg := &sync.WaitGroup{}
	for i := 0; i < 500; i++ {
		wg.Add(1)
		go func(a int) { stress("127.0.0.1:25560", wg, a, t) }(i)
	}

	wg.Wait()

}

func stress(addr string, wg *sync.WaitGroup, a int, t *testing.T) {
	c, err := net.Dial("tcp", addr)
	if err != nil {
		panic(err)
	}

	o := make([]byte, 1024)
	r := make([]byte, 1024)
	for i := 0; i < 500; i++ {
		time.Sleep(time.Millisecond * 30) // "i/o-Traffic"
		rand.Read(o)
		_, err = c.Write(o)
		if err != nil {
			panic(err)
		}
		time.Sleep(time.Millisecond * 30) // "i/o-Traffic"
		_, err = c.Read(r)
		if err != nil || !bytes.Equal(r, o) {
			panic(err)
		}
	}
	wg.Done()
}

func startEchoServer(addr string, t *testing.T) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			panic(err)
		}

		go handleRequest(conn, t)
	}
}

func handleRequest(conn net.Conn, t *testing.T) {
	defer conn.Close()
	io.Copy(conn, conn)
}
