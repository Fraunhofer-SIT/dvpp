package dns

import (
	"errors"
	"fmt"
	"net"

	mdns "github.com/miekg/dns"
)

// GetLocalNameserver return the nameserver listed in resolv.conf as a string
func GetLocalNameserver() string {
	conf, _ := mdns.ClientConfigFromFile("/etc/resolv.conf")
	return net.JoinHostPort(conf.Servers[0], conf.Port)
}

// GetAuthoritativeNameServer returns the authoritative nameserver for a given
// domain and return the result as a string and error
func GetAuthoritativeNameServer(domain string, nameserver string) (string, error) {
	c := new(mdns.Client)

	m := new(mdns.Msg)
	m.SetQuestion(mdns.Fqdn(domain), mdns.TypeNS)
	m.SetEdns0(4096, false)
	m.RecursionDesired = true

	r, _, err := c.Exchange(m, nameserver)

	if err != nil {
		return "", err
	}

	if r.Rcode != mdns.RcodeSuccess {
		return "", fmt.Errorf("received DNS response code %d", r.Rcode)
	}

	for _, a := range r.Answer {
		switch ns := a.(type) {
		case *mdns.NS:
			return ns.Ns, nil
		}
	}

	if r.Ns == nil && len(r.Ns) < 1 {
		return "", errors.New("did not receive an authoritative name server")
	}

	ns := r.Ns[0].(*mdns.SOA).Ns
	ips, err := net.LookupIP(ns)
	if err != nil {
		return "", err
	}

	for _, i := range ips {
		ip := i.To4()
		if ip != nil {
			return ip.String(), nil
		}
	}

	return "", errors.New("could not resolve A record")
}
