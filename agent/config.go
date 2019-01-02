package agent

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/dvpp/dvpp/validation/dns"
	"github.com/spf13/viper"
	"github.com/urfave/cli"
)

var paths = []string{"."}

type config interface {
	parseConfig(*cli.Context) error
}

type configAgent struct {
	ip        string
	port      string
	cert      string
	key       string
	dns       string
	users     map[string]string
	whitelist []string
}

func getFile(file string) (string, error) {
	if file[0] == '/' {
		if _, err := os.Stat(file); err == nil {
			return file, nil
		}
		return "", fmt.Errorf("Cannot find file %s.", file)
	}

	for _, p := range paths {
		c := p + "/" + file
		if _, err := os.Stat(c); err == nil {
			return c, nil
		}
	}

	return "", fmt.Errorf("Cannot find file %s.", file)
}

func (config *configAgent) parseConfig(c *cli.Context) error {
	logfile := c.String("logfile")
	if logfile != "" {
		f, err := os.OpenFile(
			logfile,
			os.O_APPEND|os.O_CREATE|os.O_RDWR,
			0666)
		if err != nil {
			log.Fatal("Could not open logfile")
		}
		log.SetOutput(f)
	}

	configFile := c.String("config")
	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.SetConfigName("agent")

		for _, p := range paths {
			viper.AddConfigPath(p)
		}
	}

	err := viper.ReadInConfig()
	if err != nil {
		return fmt.Errorf("Fatal error config file: %s \n", err)
	}

	cert, err := getFile(viper.GetString("certificate"))
	if err != nil {
		panic(err)
	}

	key, err := getFile(viper.GetString("key"))
	if err != nil {
		panic(err)
	}

	*config = configAgent{
		ip:        viper.GetString("ip"),
		port:      viper.GetString("port"),
		cert:      cert,
		key:       key,
		users:     make(map[string]string),
		whitelist: viper.GetStringSlice("whitelist"),
	}

	dnsflag := c.String("dns")
	if dnsflag != "" {
		rand.Seed(time.Now().Unix())
		ds := strings.Split(dnsflag, ",")
		config.dns = ds[rand.Intn(len(ds))]
		if !strings.Contains(config.dns, ":") {
			config.dns = config.dns + ":53"
		}
	} else {
		config.dns = dns.GetLocalNameserver()
	}

	for user, pass := range viper.GetStringMapString("users") {
		config.users[user] = pass
	}

	return nil
}
