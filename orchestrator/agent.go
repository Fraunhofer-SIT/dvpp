package orchestrator

import "fmt"

const DefaultPort = "8268"

type Agent struct {
	Name string
	Host string
	Port string
	User string
	Pass string
}

func newAgent(name string, m map[string]string) Agent {
	s := Agent{
		Name: name,
		Host: m["host"],
		Port: m["port"],
		User: m["user"],
		Pass: m["pass"],
	}

	if s.Port == "" {
		s.Port = DefaultPort
	}

	return s
}

func (a Agent) toURL() string {
	return fmt.Sprintf("https://%s:%s", a.Host, a.Port)
}
