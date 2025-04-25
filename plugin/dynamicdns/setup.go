package dynamicdns

/**
 * @author Hossein Boka <i@Ho3e.in>
 */

import (
	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
)

func init() {
	plugin.Register("dynamicdns", setup)
}

func setup(c *caddy.Controller) error {
	d := NewDynamicDNS()

	// Start REST API (from your earlier code)
	d.StartAPI()

	// Add to plugin chain
	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		d.Next = next
		return d
	})

	return nil
}

func NewDynamicDNS() *DynamicDNS {
	return &DynamicDNS{
		records: make(map[string]string),
		apiAddr: "localhost:8080", // Default
	}
}
