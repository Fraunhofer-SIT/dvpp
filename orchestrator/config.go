package orchestrator

import (
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
	"github.com/urfave/cli"
)

const (
	DefaultTimeout   = 2000
	DefaultTolerance = 0
	DefaultPoolSize  = 0
	DefaultCertsDir  = "certs"
)

var paths = []string{"."}

type config interface {
	parseConfig(*cli.Context) error
}

type configOrchestrator struct {
	XML       bool
	Timeout   int
	Tolerance int
	PoolSize  int
	CertsDir  string
	Certs     *x509.CertPool
	Agents    []Agent
	Args      []string
}

func readCert(certFile string) ([]byte, error) {
	log.Printf("Loading certificate %s.", certFile)

	cert, err := ioutil.ReadFile(certFile)
	if err != nil {
		return nil, err
	}

	return cert, nil
}

func readCerts(certsDir string) (*x509.CertPool, error) {
	certs := x509.NewCertPool()

	dir, err := os.Stat(certsDir)
	if err != nil || !dir.IsDir() {
		return nil, fmt.Errorf("cannot open directory %s", certsDir)
	}

	files, _ := ioutil.ReadDir(certsDir)
	for _, f := range files {
		name := f.Name()
		if !strings.HasPrefix(name, ".") &&
			!strings.HasPrefix(name, "_") {
			cert, err := readCert(filepath.Join(certsDir, name))
			if err != nil {
				return nil, err
			}

			certs.AppendCertsFromPEM(cert)
		}
	}

	return certs, nil
}

func (config *configOrchestrator) parseConfig(c *cli.Context) error {
	xml := c.Bool("xml")

	verbose := c.Bool("verbose")
	if !verbose {
		log.SetOutput(ioutil.Discard)
	}

	logfile := c.String("logfile")
	if logfile != "" {
		f, err := os.OpenFile(
			logfile,
			os.O_APPEND|os.O_CREATE|os.O_RDWR,
			0666)
		if err != nil {
			return err
		}
		log.SetOutput(f)
	}

	conffile := c.String("conffile")
	if conffile != "" {
		viper.SetConfigFile(conffile)
	} else {
		viper.SetConfigName("orchestrator")

		for _, p := range paths {
			viper.AddConfigPath(p)
		}

	}

	args := c.Args()
	if c.NArg() != 4 {
		return fmt.Errorf("incorrect number of arguments specified: should be cname <domain> <challenge> <response>")
	}

	viper.SetDefault("timeout", DefaultTimeout)
	viper.SetDefault("tolerance", DefaultTolerance)
	viper.SetDefault("active", DefaultPoolSize)
	viper.SetDefault("certificates", DefaultCertsDir)

	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	*config = configOrchestrator{
		XML:       xml,
		Timeout:   viper.GetInt("timeout"),
		Tolerance: viper.GetInt("tolerance"),
		PoolSize:  viper.GetInt("poolsize"),
		CertsDir:  viper.GetString("certificates"),
		Certs:     nil,
		Agents:    make([]Agent, 0),
		Args:      args,
	}

	for s1 := range viper.GetStringMapString("agents") {
		s2 := viper.GetStringMapString("agents." + s1)
		server := newAgent(s1, s2)
		config.Agents = append(config.Agents, server)
	}

	if config.PoolSize <= 0 || config.PoolSize > len(config.Agents) {
		config.PoolSize = len(config.Agents)
	}

	if float32(config.Tolerance) >= float32(config.PoolSize)/2.0 {
		return fmt.Errorf(
			"Tolerance must be smaller than 50%% of the pool size.\n"+
				"Tolerance: %d	Pool size: %d",
			config.Tolerance, config.PoolSize)
	}

	certs, err := readCerts(config.CertsDir)
	if err != nil {
		return err
	}
	config.Certs = certs

	return nil
}
