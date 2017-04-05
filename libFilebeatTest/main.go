package main

import (
	"github.com/elastic/beats/filebeat/beater"
	cfg "github.com/elastic/beats/filebeat/config"
	"github.com/elastic/beats/filebeat/crawler"
	"github.com/elastic/beats/filebeat/registrar"
	"github.com/elastic/beats/libbeat/common"
	"github.com/tdelacour/libFilebeatTest/mspooler"

	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var config = &cfg.Config{
	RegistryFile:    "registry",
	IdleTimeout:     time.Minute,
	SpoolSize:       10,
	ShutdownTimeout: time.Minute,
}

func main() {
	if len(os.Args) < 1 {
		fmt.Printf("please provide glob path to logfiles. Exiting.\n")
		os.Exit(1)
	}

	if err := initConfig(); err != nil {
		os.Exit(1)
	}

	done := make(chan struct{})

	// Shut down apparatus
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Printf("Exiting cleanly because of SIGTERM\n")
		close(done)
	}()

	run(done)
}

func initConfig() error {
	var err error
	config.ProspectorReload, err = common.NewConfigFrom(map[string]interface{}{
		"enabled": false,
	})
	if err != nil {
		fmt.Printf("Unable to create ProspectorReload config field: %v\n", err)
		return err
	}

	prospectorConfig, err := common.NewConfigFrom(map[string]interface{}{
		"enabled":    true,
		"input_type": "log",
		"paths":      []string{"/data/cloudp21328*/log"},
	})
	if err != nil {
		fmt.Printf("Unable to create Prospectors config field: %v\n", err)
		return err
	}

	config.Prospectors = []*common.Config{prospectorConfig}

	return nil
}

func run(done chan struct{}) {
	waitFinished := beater.NewSignalWait()
	waitEvents := beater.NewSignalWait()

	wgEvents := &sync.WaitGroup{}
	finishedLogger := beater.NewFinishedLogger(wgEvents)

	registrar, err := registrar.New(config.RegistryFile, finishedLogger)
	if err != nil {
		fmt.Printf("Unable to initialize registrar: %v\n", err)
		os.Exit(1)
	}

	registrarChannel := beater.NewRegistrarLogger(registrar)

	mSpooler, err := mspooler.New(registrarChannel, config)
	if err != nil {
		fmt.Printf("Unable to initialize mongo-automation spooler: %v\n", err)
		os.Exit(1)
	}

	crawler, err := crawler.New(mspooler.NewMspoolerOutlet(done, mSpooler, wgEvents), config.Prospectors, done, false)
	if err != nil {
		fmt.Printf("Unable to initialize crawler: %v\n", err)
		os.Exit(1)
	}

	err = registrar.Start()
	if err != nil {
		fmt.Printf("Unable to start registrar: %v\n", err)
		os.Exit(1)
	}
	defer registrar.Stop()

	mSpooler.Start()
	defer func() {
		waitEvents.Wait()
		registrarChannel.Close()
		mSpooler.Stop()
	}()

	err = crawler.Start(registrar, config.ProspectorReload)
	if err != nil {
		fmt.Printf("Unable to start crawler: %v\n", err)
		os.Exit(1)
	}

	waitFinished.AddChan(done)
	waitFinished.Wait()

	crawler.Stop()

	timeout := config.ShutdownTimeout
	if timeout > 0 {
		waitEvents.Add(beater.WithLog(wgEvents.Wait,
			"Continue shutdown: All enqueued events being published."))
		waitEvents.Add(beater.WithLog(beater.WaitDuration(timeout),
			"Continue shutdown: Time out waiting for events being published."))
	}

	return
}
