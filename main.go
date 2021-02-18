package main

import (
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/worldOneo/glass-proxy/cmd"
	"github.com/worldOneo/glass-proxy/cmds"
	"github.com/worldOneo/glass-proxy/config"
	"github.com/worldOneo/glass-proxy/tcpproxy"
)

//
const (
	ConfigPath = "glass.proxy.json"
)

func main() {
	cnf := loadConfig()
	bootProxy(cnf)
}

func bootProxy(cnf *config.Config) {
	rand.Seed(time.Now().UnixNano())

	proxyService := tcpproxy.NewProxyService(cnf)

	commandHandler := cmd.NewCommandHandler()
	registerCommands(commandHandler, proxyService)

	go start(proxyService)
	go healthCheck(proxyService)
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

func toConn(proxyService *tcpproxy.ProxyService, conn net.Conn) {
	host := proxyService.GetHost()

	defer func() {
		if proxyService.Config.LogConfig.LogDisconnect {
			log.Printf("%s Disconnected", conn.RemoteAddr())
		}
		conn.Close()
	}()

	if host == nil {
		log.Printf("No healthy host available.")
		return
	}
	serverConn, err := proxyService.Dial("tcp", host.Addr)
	if err != nil {
		log.Printf("Couldn't connect to %s (%s) \"%v\"", host.Name, host.Addr, err)
		conn.Close()
		return
	}

	if proxyService.Config.LogConfig.LogConnections {
		log.Printf("%s Connected to %s (%s) over %s", conn.RemoteAddr(), host.Name, host.Addr, serverConn.LocalAddr())
	}

	host.AddReverseProxy(conn, serverConn)
}

func healthCheck(proxyService *tcpproxy.ProxyService) {
	for {
		proxyService.HealthCheck()
		time.Sleep(time.Duration(proxyService.Config.HealthCheckTime) * time.Second)
	}
}

func start(proxyService *tcpproxy.ProxyService) {
	ln, err := net.Listen("tcp", proxyService.Config.Addr)
	if err != nil {
		log.Fatalf("Couldn't start the server: %v", err)
	}
	log.Printf("Listening on %s", proxyService.Config.Addr)
	for {
		conn, _ := ln.Accept()
		go toConn(proxyService, conn)
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
