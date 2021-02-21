package main

import (
	"errors"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/worldOneo/glass-proxy/cmd"
	"github.com/worldOneo/glass-proxy/cmds"
	"github.com/worldOneo/glass-proxy/config"
	"github.com/worldOneo/glass-proxy/proxy"
	"github.com/worldOneo/glass-proxy/tcp"
	"github.com/worldOneo/glass-proxy/udp"
)

//
const (
	ConfigPath = "glass.proxy.json"
)

func main() {
	cnf := loadConfig()
	rand.Seed(time.Now().UnixNano())
	bootProxy(cnf)
}

func bootProxy(cnf *config.Config) {
	var service proxy.Service
	switch strings.ToLower(cnf.Protocol) {
	case "udp", "udp4", "udp6":
		log.Printf("Starting UDP (%s) proxy on %s...", cnf.Protocol, cnf.Addr)
		udpService := udp.NewService(cnf)
		go udpService.Run()
		service = udpService
	case "tcp", "tcp4", "tcp6":
		log.Printf("Starting TCP (%s) proxy on %s...", cnf.Protocol, cnf.Addr)
		tcpService := tcp.NewProxyService(cnf)
		go tcpService.Run()
		service = tcpService
	default:
		log.Fatal(errors.New("invalid protocol. supported: tcp,udp"))
	}

	handler := cmd.NewCommandHandler()
	handler.Register("add", cmds.NewAddCommand(service).Handle)
	handler.Register("rem", cmds.NewRemCommand(service).Handle)
	handler.Register("list", cmds.NewListCommand(service).Handle)
	handler.Register("save", cmds.NewSaveCommand(service, ConfigPath).Handle)

	go handler.Listen()

	hold()
	if service.GetConfig().SaveConfigOnClose {
		log.Println("Saving config...")
		config.Create(ConfigPath, service.GetConfig())
	}
	log.Println("Stoping...")
	return
}

func hold() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
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
