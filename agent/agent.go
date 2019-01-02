package agent

import (
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/dvpp/dvpp/validation/dns"
	"github.com/dvpp/dvpp/validation/messages"
	"github.com/julienschmidt/httprouter"
	"github.com/urfave/cli"
)

type AgentInterface interface {
	isIPAllowed(string) bool
	handleCNAME(map[string]string) messages.Result
	makeHandler(func(map[string]string) messages.Result, string) httprouter.Handle
}

type Agent struct {
	*http.Server
	*httprouter.Router
	*configAgent
}

func StartAgent(c *cli.Context) error {
	var config configAgent
	err := config.parseConfig(c)
	if err != nil {
		return err
	}

	srv := newAgent(&config)
	// CNAME validation handler
	srv.GET("/cname/:Domain/:Challenge/:Response",
		srv.makeHandler(srv.handleCNAME, "CNAME"))

	log.Printf("Listening on %s.", srv.Addr)
	log.Printf("Using nameserver %s.", config.dns)
	return srv.ListenAndServeTLS(config.cert, config.key)
}

type AuthCredentials struct {
	username string
	password string
}

func newAgent(config *configAgent) Agent {
	addr := config.ip + ":" + config.port

	router := httprouter.New()

	cfg := &tls.Config{
		MinVersion: tls.VersionTLS12,
		CurvePreferences: []tls.CurveID{tls.CurveP521, tls.CurveP384,
			tls.CurveP256},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
		},
	}

	return Agent{
		&http.Server{
			Addr:      addr,
			Handler:   router,
			TLSConfig: cfg,
			TLSNextProto: make(map[string]func(*http.Server, *tls.Conn,
				http.Handler), 0),
		},
		router,
		config}
}

func (agent *Agent) isIPAllowed(ip string) bool {
	allowed := false
	myAddr := net.ParseIP(ip)

	if len(agent.whitelist) == 0 {
		allowed = true
	} else {
		for _, i := range agent.whitelist {
			if ip == i {
				allowed = true
				break
			}

			_, cidr, _ := net.ParseCIDR(i)
			if cidr != nil && cidr.Contains(myAddr) {
				allowed = true
				break
			}
		}
	}

	return allowed
}

func (agent *Agent) handleCNAME(ps map[string]string) messages.Result {
	return dns.CNAMEValidation(dns.CNAMERequest{
		Domain:    ps["Domain"],
		Challenge: ps["Challenge"],
		Response:  ps["Response"],
	}, agent.dns)
}

func (agent *Agent) makeHandler(f func(map[string]string) messages.Result, name string) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		var ip string
		w.Header().Set("Strict-Transport-Security",
			"max-age=63072000; includeSubDomains")

		user, pass, hasAuth := r.BasicAuth()
		parts := strings.Split(r.RemoteAddr, ":")
		if len(parts) == 2 {
			ip = parts[0]
		} else {
			ip = ""
		}

		if !agent.isIPAllowed(ip) {
			http.Error(w, http.StatusText(http.StatusForbidden),
				http.StatusForbidden)
			log.Printf("IP %s rejected", ip)
			return
		}

		reqPass, ok := agent.users[user]
		if len(agent.users) == 0 || (hasAuth && ok && pass == reqPass) {
			w.Header().Set("Content-Type", "application/json")
			params := make(map[string]string)
			for _, p := range ps {
				params[p.Key] = p.Value
			}
			resp := f(params)
			log.Printf("%s request from %s (%s) for %s",
				name, ip, user, params["Domain"])
			w.Write(resp.ToJSON())
		} else {
			log.Printf("Rejected access from %s (%s)", ip, user)
			w.Header().Set("WWW-Authenticate", "Basic realm=Restricted")
			http.Error(w, http.StatusText(http.StatusUnauthorized),
				http.StatusUnauthorized)
		}
	}
}
