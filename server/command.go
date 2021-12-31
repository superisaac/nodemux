package server

import (
	"context"
	"flag"
	log "github.com/sirupsen/logrus"
	"github.com/superisaac/nodeb/balancer"
	"github.com/superisaac/nodeb/chains"
	"os"
	"time"
)

func setupLogger() {
	log.SetFormatter(&log.JSONFormatter{
		TimestampFormat: time.RFC3339Nano,
	})

	logOutput := os.Getenv("LOG_OUTPUT")
	if logOutput == "" || logOutput == "console" || logOutput == "stdout" {
		log.SetOutput(os.Stdout)
	} else if logOutput == "stderr" {
		log.SetOutput(os.Stderr)
	} else {
		file, err := os.OpenFile(logOutput, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			panic(err)
		}
		log.SetOutput(file)
	}

	envLogLevel := os.Getenv("LOG_LEVEL")
	switch envLogLevel {
	case "DEBUG":
		log.SetLevel(log.DebugLevel)
	case "INFO":
		log.SetLevel(log.InfoLevel)
	case "WARN":
		log.SetLevel(log.WarnLevel)
	case "ERROR":
		log.SetLevel(log.ErrorLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}
}

func CommandStartServer() {
	serverFlags := flag.NewFlagSet("jointrpc-server", flag.ExitOnError)
	pYamlPath := serverFlags.String("f", "chains.yml", "path to chains.yml")
	pBind := serverFlags.String("b", "127.0.0.1:9000", "The http server address and port")
	pCertfile := serverFlags.String("cert", "", "tls cert file")
	pKeyfile := serverFlags.String("key", "", "tls key file")
	// parse config
	serverFlags.Parse(os.Args[1:])

	setupLogger()

	cfg := balancer.ConfigFromFile(*pYamlPath)

	b := balancer.NewBalancer()
	chains.InstallAdaptors(b)
	b.LoadFromConfig(cfg)

	balancer.SetBalancer(b)

	rootCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	b.StartSync(rootCtx)

	var httpOpts []HTTPOptionFunc
	if *pCertfile != "" && *pKeyfile != "" {
		httpOpts = append(httpOpts, WithTLS(*pCertfile, *pKeyfile))
	}

	StartHTTPServer(rootCtx, *pBind, httpOpts...)
}
