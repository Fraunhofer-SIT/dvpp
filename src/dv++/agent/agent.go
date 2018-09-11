package main

import (
	"crypto/tls"
	"dv++/validation/messages"
	"github.com/julienschmidt/httprouter"
	"log"
	"net"
	"net/http"
	"strings"
)

type AuthCredentials struct {
	username string
	password string
}

type Agent struct {
	*http.Server
	*httprouter.Router
}

func newAgent(addr string) Agent {
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

	return Agent{&http.Server{
		Addr:      addr,
		Handler:   router,
		TLSConfig: cfg,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn,
			http.Handler), 0),
	}, router}
}

func isIPAllowed(ip string) bool {
	allowed := false
	myaddr := net.ParseIP(ip)

	if len(config.whitelist) == 0 {
		allowed = true
	} else {
		for _, i := range config.whitelist {
			if ip == i {
				allowed = true
				break
			}

			_, cidr, _ := net.ParseCIDR(i)
			if cidr != nil && cidr.Contains(myaddr) {
				allowed = true
				break
			}
		}
	}

	return allowed
}

func makeHandler(
	f func(map[string]string) messages.Result, name string) httprouter.Handle {
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

		if !isIPAllowed(ip) {
			http.Error(w, http.StatusText(http.StatusForbidden),
				http.StatusForbidden)
			log.Printf("IP %s rejected", ip)
			return
		}

		reqPass, ok := config.users[user]
		if len(config.users) == 0 || (hasAuth && ok && pass == reqPass) {
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
