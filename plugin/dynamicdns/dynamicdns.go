package dynamicdns

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

/**
 * @author Hossein Boka <i@Ho3e.in>
 */

type DynamicDNS struct {
	Next    plugin.Handler
	mu      sync.RWMutex
	records map[string]string // Map of Hostname to IP address
	apiAddr string            // REST API Listen Address
}

// ServeDNS implements the plugin.Handler interface
func (d *DynamicDNS) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}
	log.Printf("HB => Query Received For: %s", state.Name())

	d.mu.RLock()
	ip, exists := d.records[state.Name()]
	d.mu.RUnlock()

	if exists {
		log.Printf("Hb => Hostname Exists And Returning A Record: %s -> %s", state.Name(), ip)

		m := new(dns.Msg)
		m.SetReply(r)
		m.Authoritative = true

		rr := &dns.A{
			Hdr: dns.RR_Header{
				Name:   state.Name(),
				Rrtype: dns.TypeA,
				Class:  dns.ClassINET,
				Ttl:    60,
			},
			A: net.ParseIP(ip),
		}
		m.Answer = append(m.Answer, rr)

		w.WriteMsg(m)
		return dns.RcodeSuccess, nil
	} else {
		log.Printf("HB => Hostname Not Exists So Chain To Another Plugin")
	}

	return plugin.NextOrFailure(d.Name(), d.Next, ctx, w, r)
}

// Name implements the Handler interface
func (d *DynamicDNS) Name() string { return "dynamicdns" }

// ## Implement REST API Handler To Manipulate ##
func (d *DynamicDNS) StartAPI() {
	http.HandleFunc("/add", d.handleAdd)
	http.HandleFunc("/remove", d.handleRemove)
	http.HandleFunc("/list", d.handleList)

	go func() {
		if err := http.ListenAndServe(d.apiAddr, nil); err != nil {
			log.Printf("Failed to start API server: %v", err)
		}
	}()
}

func (d *DynamicDNS) handleAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	hostname := r.FormValue("hostname")
	ip := r.FormValue("ip")

	if hostname == "" || ip == "" {
		http.Error(w, "hostname and ip are required", http.StatusBadRequest)
		return
	}

	d.mu.Lock()
	d.records[hostname] = ip
	d.mu.Unlock()

	w.WriteHeader(http.StatusOK)
}

func (d *DynamicDNS) handleRemove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	hostname := r.FormValue("hostname")

	d.mu.Lock()
	delete(d.records, hostname)
	d.mu.Unlock()

	w.WriteHeader(http.StatusOK)
}

func (d *DynamicDNS) handleList(w http.ResponseWriter, r *http.Request) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	json.NewEncoder(w).Encode(d.records)
}
