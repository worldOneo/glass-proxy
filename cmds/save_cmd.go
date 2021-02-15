package cmds

import (
	"github.com/worldOneo/glass-proxy/config"
	"github.com/worldOneo/glass-proxy/tcpproxy"
)

// SaveCmd saves the config
type SaveCmd struct {
	proxyService *tcpproxy.ProxyService
	cnf          string
}

// NewSaveCommand creates a new save cmf
func NewSaveCommand(proxyService *tcpproxy.ProxyService, confPath string) *SaveCmd {
	return &SaveCmd{
		cnf:          confPath,
		proxyService: proxyService,
	}
}

// Handle saves the config
func (s *SaveCmd) Handle(args []string) {
	config.Create(s.cnf, s.proxyService.Config)
}
