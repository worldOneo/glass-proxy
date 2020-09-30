package main

import (
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
	"github.com/worldOneo/glass-proxy/src/tcpproxy"
)

//
const (
	ConfigPath = "glass.proxy.json"
)

var proxyService *tcpproxy.ProxyService

func main() {
	cnf := loadConfig()
	rand.Seed(time.Now().UnixNano())

	proxyService = &tcpproxy.ProxyService{
		Config:         cnf,
		CommandHandler: cmd.NewCommandHandler(),
	}
	proxyService.LoadHosts()

	commandHandler := cmd.NewCommandHandler()
	registerCommands(commandHandler, proxyService)

	go start()
	go healthCheck()
	go commandHandler.Listen()

	hold()
	stop(proxyService)
	log.Println("Stoping...")
	return
}

func hold() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}

func stop(proxyService *tcpproxy.ProxyService) {
	if proxyService.Config.SaveConfigOnClose {
		log.Println("Saving config...")
		config.Create(ConfigPath, proxyService.Config)
	}
}

func toConn(conn net.Conn) {
	host := proxyService.GetHost()
	if host == nil {
		conn.Close()
		return
	}
	serverConn, err := proxyService.Dial("tcp", host.Addr)
	if err != nil {
		log.Printf("Couldn't connect to %s (%s) \"%v\"", host.Name, host.Addr, err)
		conn.Close()
		return
	}
	reverseProxy := tcpproxy.NewReverseProxy(conn, serverConn)
	reverseProxy.Pipe()

	if proxyService.Config.LoggConnections {
		log.Printf("%s Connected to %s (%s) over %s", conn.RemoteAddr(), host.Name, host.Addr, serverConn.LocalAddr())
	}
}

func healthCheck() {
	for {
		proxyService.HealthCheck()
		time.Sleep(time.Duration(proxyService.Config.HealthCheckTime) * time.Second)
	}
}

func start() {
	ln, err := net.Listen("tcp", proxyService.Config.Addr)
	if err != nil {
		log.Fatalf("Couldn't start the server: %v", err)
	}
	log.Printf("Listening on %s", proxyService.Config.Addr)
	for {
		conn, _ := ln.Accept()
		go toConn(conn)
	}
}

func registerCommands(cmdHandler *cmd.CommandHandler, proxyService *tcpproxy.ProxyService) {
	cmdHandler.Register("add", cmds.NewAddCommand(proxyService).Handle)
	cmdHandler.Register("rem", cmds.NewRemCommand(proxyService).Handle)
	cmdHandler.Register("list", cmds.NewListCommand(proxyService).Handle)
	cmdHandler.Register("save", cmds.NewSaveCommand(proxyService, ConfigPath).Handle)
}

func loadConfig() *config.Config {
	cnf, cnfErr := config.Load(ConfigPath)

	if cnfErr != nil {
		cnf = config.Default()
		creErr := config.Create(ConfigPath, cnf)
		if creErr != nil {
			log.Fatal("Couldn't load the Config")
		}
	}
	return cnf
}
