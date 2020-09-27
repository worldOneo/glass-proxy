package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/worldOneo/glass-proxy/src/cmd"
	"github.com/worldOneo/glass-proxy/src/cmds"
	"github.com/worldOneo/glass-proxy/src/config"
	"github.com/worldOneo/glass-proxy/src/handler"
	"github.com/worldOneo/glass-proxy/src/tcpproxy"
)

//
const (
	ConfigPath = "glass.proxy.json"
)

var proxyService *tcpproxy.ProxyService

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

	proxyService = &tcpproxy.ProxyService{
		Config:         cnf,
		CommandHandler: cmd.NewCommandHandler(),
	}
	proxyService.LoadHosts()

	commandHandler := cmd.NewCommandHandler()
	commandHandler.Register("add", cmds.NewAddCommand(proxyService).Handle)
	commandHandler.Register("rem", cmds.NewRemCommand(proxyService).Handle)
	commandHandler.Register("list", cmds.NewListCommand(proxyService).Handle)

	ln, err := net.Listen("tcp", proxyService.Config.Addr)
	if err != nil {
		return
	}

	log.Printf("Listening on %s", proxyService.Config.Addr)

	go healthCheck()
	go acceptConnections(ln)
	go commandHandler.Listen()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	fmt.Println("Stoping...")
	return
}

func toConn(conn net.Conn) {
	host := getHost()
	if host == nil {
		conn.Close()
		return
	}
	serverConn, err := net.Dial("tcp", host.Addr)
	if err != nil {
		log.Printf("Couldn't connect to %s (%s)", host.Name, host.Addr)
		return
	}
	reverseProxy := tcpproxy.NewReverseProxy(conn, serverConn)
	go reverseProxy.Pipe()

	if proxyService.Config.LoggConnections {
		log.Printf("%s Connected to %s (%s)", conn.RemoteAddr(), host.Name, host.Addr)
	}
}

func getHost() *handler.TCPHost {
	l := len(proxyService.Hosts)
	if l == 0 {
		return nil
	}
	i := rand.Intn(l)
	h := proxyService.Hosts[i]
	if h.Status.Online {
		return h
	}
	mx := i + l
	for j := i; j < mx; j++ {
		h = proxyService.Hosts[j%l]
		if h.Status.Online {
			return h
		}
	}
	return nil
}

func healthCheck() {
	for {
		handler.CheckHosts(proxyService.Hosts)
		time.Sleep(time.Duration(proxyService.Config.HealthCheckTime) * time.Second)
	}
}

func acceptConnections(ln net.Listener) {
	for {
		conn, _ := ln.Accept()
		go toConn(conn)
	}
}
