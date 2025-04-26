package hosthealth

import (
    "github.com/coredns/coredns/core/dnsserver"
    "github.com/coredns/caddy"
)

/**
 * @author Hossein Boka <i@Ho3e.in>
 */

 func init() {
    plugin.Register("hosthealth", setup)
}

func setup(c *caddy.Controller) error {
    hr := New()

    for c.Next() { // Loop over healthrecords blocks
        for c.NextBlock() { // Parse inside the block
            args := c.RemainingArgs()
            if len(args) != 2 {
                return c.ArgErr() // Each line must have exactly 2 arguments
            }
            domain := args[0]
            ip := args[1]

            hr.Mutex.Lock()
            hr.Records[domain] = append(hr.Records[domain], &record{
                IP:    ip,
                Alive: true, // Assume alive at startup
            })
            hr.Mutex.Unlock()
        }
    }

    dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
        hr.Next = next
        return hr
    })

    return nil
}