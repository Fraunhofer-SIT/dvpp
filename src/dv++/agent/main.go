package main

import (
	"dv++/validation/dns"
	"dv++/validation/messages"
	"log"
)

func handleCNAME(ps map[string]string) messages.Result {
	return dns.CNAMEValidation(dns.CNAMERequest{
		Domain:    ps["Domain"],
		Challenge: ps["Challenge"],
		Response:  ps["Response"],
	}, config.dns)
}

func main() {
	err := parseConfig()
	if err != nil {
		panic(err)
	}

	ipport := config.ip + ":" + config.port
	srv := newAgent(ipport)
	srv.GET("/cname/:Domain/:Challenge/:Response",
		makeHandler(handleCNAME, "CNAME"))

	log.Printf("Listening on %s.", ipport)
	log.Printf("Using nameserver %s.", config.dns)
	log.Fatal(srv.ListenAndServeTLS(config.cert, config.key))
}
