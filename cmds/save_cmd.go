package cmds

import (
	"github.com/worldOneo/glass-proxy/config"
	"github.com/worldOneo/glass-proxy/proxy"
)

// SaveCmd saves the config
type SaveCmd struct {
	proxyService proxy.Service
	cnf          string
}

// NewSaveCommand creates a new save cmf
func NewSaveCommand(proxyService proxy.Service, confPath string) *SaveCmd {
	return &SaveCmd{
		cnf:          confPath,
		proxyService: proxyService,
	}
}

// Handle saves the config
func (s *SaveCmd) Handle(args []string) {
	config.Create(s.cnf, s.proxyService.GetConfig())
}
