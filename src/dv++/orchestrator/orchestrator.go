package main

import (
	"crypto/tls"
	"dv++/validation/messages"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

type Orchestrator struct {
	*http.Client
}

func newOrchestrator() Orchestrator {
	return Orchestrator{
		&http.Client{
			Timeout: time.Millisecond * time.Duration(config.Timeout),
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs: config.Certs,
				},
			},
		},
	}
}

func (o Orchestrator) sendRequest(a Agent, r messages.HTTPRequest,
	c chan messages.Result) {
	var response messages.Result
	var resp *http.Response
	var body []byte
	msg := messages.NewResult()
	url := a.toURL() + r.ToPath()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		goto fail
	}

	if a.User != "" {
		req.SetBasicAuth(a.User, a.Pass)
	}
	resp, err = o.Do(req)
	if err != nil {
		goto fail
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		msg.AppendErrorWithPrefix(
			fmt.Sprintf("returned HTTP code %d", resp.StatusCode),
			a.Name,
		)
		c <- msg
		return
	}

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		goto fail
	}

	err = json.Unmarshal([]byte(body), &response)
	if err != nil {
		goto fail
	}

	msg.Success = response.Success
	msg.Response = response.Response
	msg.AppendErrorsWithPrefix(response.Errors, a.Name)

	c <- msg
	return

fail:
	log.Println(err.Error())
	msg.AppendErrorWithPrefix(err.Error(), a.Name)
	c <- msg

}

func getRandomAgentPool(n int) []Agent {
	if n == len(config.Agents) {
		return config.Agents
	}

	pool := []Agent{}
	rand.Seed(time.Now().UnixNano() ^ int64(os.Getpid()))

	for _, i := range rand.Perm(len(config.Agents))[:n] {
		pool = append(pool, config.Agents[i])
	}

	return pool
}

func (o Orchestrator) validateDomain(req messages.HTTPRequest) messages.Result {
	resp := messages.NewResult()
	reqs := []messages.Result{}
	values := make(map[string]int)

	pool := getRandomAgentPool(config.PoolSize)

	c := make(chan messages.Result, len(pool))

	for _, a := range pool {
		log.Printf("Fetching result from %s...\n", a.Name)
		go o.sendRequest(a, req, c)
	}

	for range pool {
		reqs = append(reqs, <-c)
	}

	for _, r := range reqs {
		if r.Success {
			values[r.Response]++
		} else {
			resp.AppendErrors(r.Errors)
			log.Println(r.Errors)
		}
	}

	response := ""
	successes := 0
	for r, c := range values {
		if c > successes {
			response = r
			successes = c
		}
	}

	failed := len(pool) - successes
	resp.Success = failed <= config.Tolerance
	if resp.Success {
		resp.Response = response
	} else if len(resp.Errors) == 0 {
		resp.AppendError(
			fmt.Sprintf("local: %d response(s) failed (tolerance = %d)",
				failed, config.Tolerance,
			),
		)
	}

	return resp
}
