package cmds

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/worldOneo/glass-proxy/proxy"
)

// ListCmd is a command to add a server to the Proxy
type ListCmd struct {
	proxyService proxy.Service
}

// NewListCommand creates a new AddCmd
func NewListCommand(proxyService proxy.Service) *ListCmd {
	return &ListCmd{
		proxyService: proxyService,
	}
}

// Handle handles the commands and list every server and their status
func (l *ListCmd) Handle(args []string) {
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 8, 8, 0, '\t', 0)
	defer w.Flush()

	fmt.Fprintf(w, "%s\t|%s\t|%s\t|%s\t|%s\t\n", "Index", "Name", "Address", "Online", "Connections")
	for i, h := range l.proxyService.ListHosts() {
		fmt.Fprintf(w, "%d\t|%s\t|%s\t|%t\t|%d\t\n", i, h.GetName(), h.GetAddr(), h.GetStatus().IsOnline(), h.GetStatus().GetConnectionCount())
	}
}
