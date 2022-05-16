package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/vatsimnerd/simwatch"
	"github.com/vatsimnerd/simwatch/config"
)

func main() {
	var configFilename string
	flag.StringVar(&configFilename, "c", "simwatch", "config name")
	flag.Parse()

	sigs := make(chan os.Signal, 1024)
	signal.Notify(sigs, syscall.SIGINT)
	defer signal.Reset(syscall.SIGINT)
	defer close(sigs)

	cfg, err := config.Read(configFilename)
	if err != nil {
		logrus.Fatal(err)
	}

	level, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)

	s := simwatch.NewServer(cfg)

	go s.Start()

	<-sigs
	logrus.Info("SIGINT caught, stopping server")

	err = s.Stop()
	if err != nil {
		logrus.WithError(err).Error("error stopping server")
	}

}
