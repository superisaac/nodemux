package server

import (
	"context"
	"flag"
	"fmt"
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"github.com/superisaac/nodepool/balancer"
	"github.com/superisaac/nodepool/chains"
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

func watchConfig(rootCtx context.Context, yamlPath string, fetch bool) {
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
				nbcfg, err := balancer.ConfigFromFile(event.Name)
				if err != nil {
					log.Warnf("error config %s", err)
				} else {
					b := balancer.BalancerFromConfig(nbcfg)
					b.StartSync(rootCtx, fetch)
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
	pYamlPath := serverFlags.String("f", "nodepool.yml", "path to nodepool.yml")
	pWatchConfig := serverFlags.Bool("w", false, "watch config changes using fsnotify")
	pFetchEndpoints := serverFlags.Bool("fetch", true, "fetch endpoints statuses")

	pServerYmlPath := serverFlags.String("server", "", "the path to server.yml")
	pBind := serverFlags.String("b", "", "The http server address and port, default is 127.0.0.1:9000")

	pLogfile := serverFlags.String("log", "", "path to log output, default is stdout")

	// parse config
	serverFlags.Parse(os.Args[1:])

	setupLogger(*pLogfile)

	// parse server.yml
	serverCfg := NewServerConfig()
	serverYamlPath := *pServerYmlPath
	if serverYamlPath != "" {
		if _, err := os.Stat(serverYamlPath); err != nil && os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "server.yml not exist\n")
			os.Exit(1)
		}

		err := serverCfg.Load(serverYamlPath)
		if err != nil {
			panic(err)
		}
		bind := *pBind
		if bind != "" {
			serverCfg.Bind = bind
		}
	}

	// parse nodepool.yml
	yamlPath := *pYamlPath
	if _, err := os.Stat(yamlPath); err != nil && os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "config yaml not exist\n")
		os.Exit(1)
	}

	nbcfg, err := balancer.ConfigFromFile(*pYamlPath)
	if err != nil {
		panic(err)
	}

	// initial delegator factory and add chains support to it
	factory := balancer.GetDelegatorFactory()
	chains.InstallAdaptors(factory)

	// initialize balancer
	b := balancer.BalancerFromConfig(nbcfg)

	balancer.SetBalancer(b)

	rootCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	b.StartSync(rootCtx, *pFetchEndpoints)

	if *pWatchConfig {
		go watchConfig(rootCtx, *pYamlPath, *pFetchEndpoints)
	}

	StartHTTPServer(rootCtx, serverCfg)
}
