package dns

import (
	"dv++/validation/messages"
	"fmt"
	mdns "github.com/miekg/dns"
	"net"
)

type CNAMERequest struct {
	Domain    string
	Challenge string
	Response  string
}

// ToPath returns the Request as an URI path that can be used in HTTP GET
// requests.
func (r CNAMERequest) ToPath() string {
	return fmt.Sprintf("/cname/%s/%s/%s", r.Domain, r.Challenge, r.Response)
}

// Validation performs the CNAME validation using the supplied nameserver,
// challenge and response.
// It returns an messages.Data object.
func CNAMEValidation(req CNAMERequest, nameserver string) messages.Result {
	req.Response = mdns.Fqdn(req.Response)
	fmt.Println(nameserver)
	ns, err := GetAuthoritativeNameServer(req.Domain, nameserver)
	if err != nil {
		return messages.Result{Success: false, Errors: []string{err.Error()}}
	}

	c := new(mdns.Client)

	m := new(mdns.Msg)
	m.SetQuestion(mdns.Fqdn(req.Challenge), mdns.TypeCNAME)
	m.RecursionDesired = true

	r, _, err := c.Exchange(m, net.JoinHostPort(ns, "53"))

	if err != nil {
		return messages.Result{
			Success: false,
			Errors:  []string{err.Error()},
		}
	}

	if r.Rcode != mdns.RcodeSuccess {
		return messages.Result{
			Success: false,
			Errors: []string{
				fmt.Sprintf("Received DNS response code %d", r.Rcode),
			},
		}
	}

	for _, a := range r.Answer {
		switch c := a.(type) {
		case *mdns.CNAME:
			resp := messages.NewResult()
			resp.Success = c.Target == req.Response
			resp.Response = c.Target
			if !resp.Success {
				resp.AppendError("Invalid response")
			}

			return resp
		}
	}

	return messages.Result{Success: false, Errors: []string{"No response"}}
}
