package proxy

import "github.com/worldOneo/glass-proxy/config"

type Service interface {
	AddHost(config.HostConfig)
	RemHost(string)
	GetConfig() *config.Config
	ListHosts() []Host
}

type Host interface {
	GetName() string
	GetAddr() string
	GetStatus() HostStatus
}

type HostStatus interface {
	IsOnline() bool
}
