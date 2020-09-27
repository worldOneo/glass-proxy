package main

import (
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/worldOneo/glass-proxy/src/config"
	"github.com/worldOneo/glass-proxy/src/handler"
	"github.com/worldOneo/glass-proxy/src/tcpproxy"
)

// ProxyService with everything we need
type ProxyService struct {
	hosts  []*handler.TCPHost
	config *config.Config
}

//
const (
	ConfigPath = "glass.proxy.json"
)

var proxyService *ProxyService

func main() {
	cnf, cnfErr := config.Load(ConfigPath)

	if cnfErr != nil {
		cnf = config.Default()
		creErr := config.Create(ConfigPath, cnf)
		if creErr != nil {
			log.Fatal("Couldn't load the Config")
		}
	}
	rand.Seed(time.Now().UnixNano())

	hosts := make([]*handler.TCPHost, 0)
	for _, host := range cnf.Hosts {
		newHost := handler.NewTCPHost(host.Name, host.Addr)
		hosts = append(hosts, newHost)
	}

	proxyService = &ProxyService{
		config: cnf,
		hosts:  hosts,
	}

	ln, err := net.Listen("tcp", proxyService.config.Addr)
	if err != nil {
		return
	}

	log.Printf("Listening on %s", proxyService.config.Addr)
	go healthCheck()
	for {
		conn, _ := ln.Accept()
		go toConn(conn)
	}
}

func toConn(conn net.Conn) {
	host := getHost()
	serverConn, err := net.Dial("tcp", host.Addr)
	if err != nil {
		log.Printf("Couldn't connect to %s (%s)", host.Name, host.Addr)
		return
	}
	reverseProxy := tcpproxy.NewReverseProxy(conn, serverConn)
	go reverseProxy.Pipe()

	if proxyService.config.LoggConnections {
		log.Printf("%s Connected to %s (%s)", conn.RemoteAddr(), host.Name, host.Addr)
	}
}

func getHost() *handler.TCPHost {
	l := len(proxyService.hosts)
	i := rand.Intn(l)
	h := proxyService.hosts[i]
	if h.Status.Online {
		return h
	}
	mx := i + l
	for j := i; j < mx; j++ {
		h = proxyService.hosts[j%l]
		if h.Status.Online {
			return h
		}
	}
	return nil
}

func healthCheck() {
	for {
		time.Sleep(time.Duration(proxyService.config.HealthCheckTime) * time.Minute)
		handler.CheckHosts(proxyService.hosts)
	}
}
