package hosthealth

import (
    "context"
    "net"
    "sync"
    "time"

    "github.com/coredns/coredns/plugin"
    "github.com/miekg/dns"
)

/**
 * @author Hossein Boka <i@Ho3e.in>
 */

type record struct {
    IP    string
    Alive bool
}

type HealthRecords struct {
    Next    plugin.Handler
    Records map[string][]*record
    Mutex   sync.RWMutex
}

func New() *HealthRecords {
    hr := &HealthRecords{
        Records: make(map[string][]*record),
    }

    go hr.runHealthChecks()
    return hr
}

func (hr *HealthRecords) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
    q := r.Question[0]

    if q.Qtype != dns.TypeA {
        return plugin.NextOrFailure(hr.Name(), hr.Next, ctx, w, r)
    }

    fqdn := q.Name

    hr.Mutex.RLock()
    recs, found := hr.Records[fqdn]
    hr.Mutex.RUnlock()

    if !found {
        return plugin.NextOrFailure(hr.Name(), hr.Next, ctx, w, r)
    }

    var answers []dns.RR
    for _, rec := range recs {
        if rec.Alive {
            answers = append(answers, &dns.A{
                Hdr: dns.RR_Header{Name: fqdn, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60},
                A:   net.ParseIP(rec.IP),
            })
        }
    }

    if len(answers) == 0 {
        return dns.RcodeNameError, nil // NXDOMAIN
    }

    m := new(dns.Msg)
    m.SetReply(r)
    m.Answer = answers
    w.WriteMsg(m)
    return dns.RcodeSuccess, nil
}

func (hr *HealthRecords) Name() string { 
	return "hosthealth" 
}
