package server

import (
	"context"
	"flag"
	"github.com/superisaac/nodeb/balancer"
	"github.com/superisaac/nodeb/chains"
	"os"
)

func CommandStartServer() {
	serverFlags := flag.NewFlagSet("jointrpc-server", flag.ExitOnError)
	pYamlPath := serverFlags.String("f", "balancer.yml", "path to balancer.yml")
	pBind := serverFlags.String("b", "127.0.0.1:9000", "The http server address and port")
	pCertfile := serverFlags.String("cert", "", "tls cert file")
	pKeyfile := serverFlags.String("key", "", "tls key file")

	// parse config
	serverFlags.Parse(os.Args[1:])

	cfg := balancer.ConfigFromFile(*pYamlPath)

	b := balancer.NewBalancer()
	b.LoadFromConfig(cfg)
	chains.InstallAdaptors(b)

	rootCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	rootCtx = b.BindToContext(rootCtx)

	var httpOpts []HTTPOptionFunc
	if *pCertfile != "" && *pKeyfile != "" {
		httpOpts = append(httpOpts, WithTLS(*pCertfile, *pKeyfile))
	}

	StartHTTPServer(rootCtx, *pBind, httpOpts...)

}
