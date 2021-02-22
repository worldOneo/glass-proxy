package proxy

import "github.com/worldOneo/glass-proxy/config"

// Service defines a service interface with the abillity to
// Add/Get/Remove hosts and get its config
type Service interface {
	AddHost(config.HostConfig)
	RemHost(string)
	GetConfig() *config.Config
	ListHosts() []Host
}

// Host basic configuration for a host
// a host holds its own name, address and status
type Host interface {
	GetName() string
	GetAddr() string
	GetStatus() HostStatus
}

// HostStatus enable lookups on dynamic information about a host.
type HostStatus interface {
	IsOnline() bool
	GetConnectionCount() int
}
