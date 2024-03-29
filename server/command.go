package server

import (
	"context"
	"flag"
	"fmt"
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"github.com/superisaac/nodemux/chains"
	"github.com/superisaac/nodemux/core"
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

func watchConfig(rootCtx context.Context, configPath string, fetch bool) {
	log.Infof("watch the config %s", configPath)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer watcher.Close()

	err = watcher.Add(configPath)
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
				log.Infof("watch config, file %s changed, event %#v", configPath, event)
				nbcfg, err := nodemuxcore.ConfigFromFile(event.Name)
				if err != nil {
					log.Warnf("error config %s", err)
				} else {
					b := nodemuxcore.MultiplexerFromConfig(nbcfg)
					b.StartSync(rootCtx, fetch)
					time.Sleep(1 * time.Second)
					nodemuxcore.SetMultiplexer(b)

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
	serverFlags := flag.NewFlagSet("nodemux", flag.ExitOnError)
	pConfigPath := serverFlags.String("f", "nodemux.yaml", "path to nodemux.yml or nodemux.json")
	pWatchConfig := serverFlags.Bool("w", false, "watch config changes using fsnotify")
	pNoSyncEndpoints := serverFlags.Bool("nosync", false, "sync endpoints statuses")

	pServerConfigPath := serverFlags.String("server", "", "the path to server.yml or server.json")
	pBind := serverFlags.String("b", "", "The http server address and port, default is 127.0.0.1:9000")

	pMetricsBind := serverFlags.String("metrics-bind", "", "The metrics server host and port")

	pLogfile := serverFlags.String("log", "", "path to log output, default is stdout")

	// parse config
	serverFlags.Parse(os.Args[1:])

	setupLogger(*pLogfile)

	// parse server.yml
	serverCfg := NewServerConfig()
	serverConfigPath := *pServerConfigPath
	if serverConfigPath != "" {
		if _, err := os.Stat(serverConfigPath); err != nil && os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "server config file not exist\n")
			os.Exit(1)
		}

		err := serverCfg.Load(serverConfigPath)
		if err != nil {
			panic(err)
		}
		bind := *pBind
		if bind != "" {
			serverCfg.Bind = bind
		}
	}

	if *pMetricsBind != "" {
		serverCfg.Metrics.Bind = *pMetricsBind
	}

	// parse nodemux.yml
	configPath := *pConfigPath
	if _, err := os.Stat(configPath); err != nil && os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "config file not exist\n")
		os.Exit(1)
	}

	nbcfg, err := nodemuxcore.ConfigFromFile(configPath)
	if err != nil {
		panic(err)
	}

	// initial delegator factory and add chains support to it
	factory := nodemuxcore.GetDelegatorFactory()
	chains.InstallAdaptors(factory)

	// initialize nodemuxcore
	b := nodemuxcore.MultiplexerFromConfig(nbcfg)

	nodemuxcore.SetMultiplexer(b)

	rootCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nosync := *pNoSyncEndpoints
	b.StartSync(rootCtx, !nosync)

	if *pWatchConfig {
		go watchConfig(rootCtx, configPath, !nosync)
	}

	StartHTTPServer(rootCtx, serverCfg)
}
