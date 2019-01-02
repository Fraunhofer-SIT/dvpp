package main

import (
	"github.com/dvpp/dvpp/agent"
	"github.com/dvpp/dvpp/orchestrator"
	"github.com/urfave/cli"
	"log"
	"os"
)

var (
	version = "dev"
)

var dnsServersFlag = cli.StringFlag{
	Name:  "dns, d",
	Usage: "DNS servers",
}
var configFileFlag = cli.StringFlag{
	Name:  "config, c",
	Usage: "Load configAgent file from `FILE`",
}
var logFileFlag = cli.StringFlag{
	Name:  "logfile, l",
	Usage: "Load log file from `FILE`",
}
var verbosityFlag = cli.BoolFlag{
	Name:  "verbose, v",
	Usage: "Verbose output",
}
var outputXMLFlag = cli.BoolFlag{
	Name:  "xml, x",
	Usage: "XML output",
}

func main() {
	app := cli.NewApp()
	app.Version = version
	app.Usage = "domain validation ++"
	// =====Commands=====
	app.Commands = []cli.Command{
		cli.Command{
			Name:  "agent",
			Usage: "Run an agent that performs the validation of a domain requested by the orchestrator.",
			Flags: toArray(dnsServersFlag, configFileFlag, logFileFlag),
			Action: func(c *cli.Context) error {
				return agent.StartAgent(c)
			},
		},
		cli.Command{
			Name: "orchestrator",
			Usage: "Run the orchestrator that coordinates the validation by sending validation requests to all " +
				"agents and verifying the result",
			Flags:     toArray(configFileFlag, logFileFlag, verbosityFlag, outputXMLFlag),
			UsageText: "main orchestrator [command options] cname <domain> <challenge> <response>",
			Action: func(c *cli.Context) error {
				return orchestrator.StartOrchestrator(c)
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func toArray(flags ...cli.Flag) []cli.Flag {
	return flags
}
