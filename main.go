package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/Sirupsen/logrus"
)

var (
	configFile = flag.String("config-file", "./config.yaml", "Configuration file")
	errChan    = make(chan error, 10)
	signalChan = make(chan os.Signal, 1)
)

func init() {
	flag.Parse()
}

func main() {
	config, err := ConfigFromFile(*configFile)
	if err != nil {
		logrus.Fatal(err)
	}

	server := NewServer(config)
	if err := server.Init(); err != nil {
		logrus.Fatal(err)
	}

	go func() {
		errChan <- server.Start()
	}()

	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case err := <-errChan:
			if err != nil {
				logrus.Fatal(err)
			}
		case signal := <-signalChan:
			logrus.Infof("Captured %v. Exiting...", signal)
			if err := server.Stop(); err != nil {
				logrus.Fatal(err)
			}
			logrus.Info("Bye")
			os.Exit(0)
		}
	}

}
