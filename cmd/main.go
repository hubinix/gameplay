// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
// ------------------------------------------------------------

package main

import (
	"github.com/hubinix/gameplay/internal/http"
	"runtime/pprof"

	"github.com/jessevdk/go-flags"

	"github.com/hubinix/gameplay/internal/db"
	"github.com/hubinix/gameplay/pkg/logger"
	"github.com/hubinix/gameplay/pkg/signals"
	"os"
	"os/signal"

	"syscall"
	"time"
)

var log = logger.NewLogger("dapr.sentry")

// Opts represents command line options
type Opts struct {
	BindAddress string `long:"bind" env:"BIND" default:"127.0.0.1:3000" description:"ip:port to bind for a node"`
	JoinAddress string `long:"join" env:"JOIN" default:"" description:"ip:port to join for a node"`
	Bootstrap   bool   `long:"bootstrap" env:"BOOTSTRAP" description:"bootstrap a cluster"`
	DataDir     string `long:"datadir" env:"DATA_DIR" default:"/tmp/data/" description:"Where to store system data"`
}

const (
	defaultCredentialsPath = "/var/run/dapr/credentials"
	// defaultDaprSystemConfigName is the default resource object name for Dapr System Config
	defaultDaprSystemConfigName = "daprsystem"

	profFile    = "./frontier-meta.prof"
	healthzPort = 8087
)

func main() {

	var opts Opts
	p := flags.NewParser(&opts, flags.Default)
	if _, err := p.ParseArgs(os.Args[1:]); err != nil {
		log.Info(err)
	}

	log.Info("[INFO] '%s' is used to store files of the node", opts.DataDir)
	f, err := os.OpenFile(profFile, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Errorf("cpu prof err")
	}
	defer f.Close()
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	ctx := signals.Context()

	go func() {
		httpServer := http.NewServer(log)
		httpServer.Ready()

		err := httpServer.Run(ctx, healthzPort)
		if err != nil {
			log.Fatalf("failed to start healthz server: %s", err)
		}
	}()
	db.GetDynamodbManager().Start(ctx)
	<-stop
	shutdownDuration := 5 * time.Second
	log.Infof("allowing %s for graceful shutdown to complete", shutdownDuration)
	<-time.After(shutdownDuration)
}
