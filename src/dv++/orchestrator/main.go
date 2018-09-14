package main

import (
	"dv++/validation/dns"
	"fmt"
	"os"
)

func main() {
	err := parseConfig()
	if err != nil {
		panic(err)
	}

	req := dns.CNAMERequest{
		Domain:    config.Args[1],
		Challenge: config.Args[2],
		Response:  config.Args[3],
	}

	c := newOrchestrator()

	resp := c.validateDomain(req)

	if config.XML {
		fmt.Println(string(resp.ToXML()))
	} else {
		fmt.Println(string(resp.ToJSON()))
	}

	if !resp.Success {
		os.Exit(1)
	}
}
