package server

import (
	"context"
	"flag"
	"fmt"
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"github.com/superisaac/nodeb/balancer"
	"github.com/superisaac/nodeb/cfg"
	"github.com/superisaac/nodeb/chains"
	"os"
	"time"
)

func setupLogger(logOutput string) {
	log.SetFormatter(&log.JSONFormatter{
		TimestampFormat: time.RFC3339Nano,
	})

	if logOutput == "" {
		logOutput = os.Getenv("LOG_OUTPUT")
	}
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

func watchConfig(rootCtx context.Context, yamlPath string) {
	log.Infof("watch the config %s", yamlPath)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer watcher.Close()

	err = watcher.Add(yamlPath)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(rootCtx)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			log.Debugf("config watcher done")
			return
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			if event.Op&fsnotify.Write == fsnotify.Write {
				log.Infof("watch config, file %s changed, event %#v", yamlPath, event)
				cfg, err := cfg.ConfigFromFile(event.Name)
				if err != nil {
					log.Warnf("error config %s", err)
				} else {
					b := balancer.NewBalancer()
					b.LoadFromConfig(cfg.Endpoints)
					b.StartSync(rootCtx)
					time.Sleep(1 * time.Second)
					balancer.SetBalancer(b)

				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Warnf("config watch error %s", err)
		}

	}
}

func CommandStartServer() {
	serverFlags := flag.NewFlagSet("jointrpc-server", flag.ExitOnError)
	pYamlPath := serverFlags.String("f", "nodeb.yml", "path to nodeb.yml")
	pWatchConfig := serverFlags.Bool("w", false, "watch config changes using fsnotify")

	pBind := serverFlags.String("b", "127.0.0.1:9000", "The http server address and port")
	pCertfile := serverFlags.String("cert", "", "tls cert file")
	pKeyfile := serverFlags.String("key", "", "tls key file")

	pLogfile := serverFlags.String("log", "", "path to log output, default is stdout")

	// parse config
	serverFlags.Parse(os.Args[1:])

	setupLogger(*pLogfile)

	yamlPath := *pYamlPath

	if _, err := os.Stat(yamlPath); err != nil && os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "config yaml not exist\n")
		os.Exit(1)
	}

	cfg, err := cfg.ConfigFromFile(*pYamlPath)
	if err != nil {
		panic(err)
	}

	// initial delegator factory and add chains support to it
	factory := balancer.GetDelegatorFactory()
	chains.InstallAdaptors(factory)

	// initialize balancer
	b := balancer.NewBalancer()
	b.LoadFromConfig(cfg.Endpoints)

	balancer.SetBalancer(b)

	rootCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	b.StartSync(rootCtx)

	if *pWatchConfig {
		go watchConfig(rootCtx, *pYamlPath)
	}

	var httpOpts []HTTPOptionFunc
	if *pCertfile != "" && *pKeyfile != "" {
		httpOpts = append(httpOpts, WithTLS(*pCertfile, *pKeyfile))
	}
	StartHTTPServer(rootCtx, *pBind, httpOpts...)
}
