package main

import (
	"github.com/anonopiran/Fly2Stats/internal/config"
	"github.com/anonopiran/Fly2Stats/internal/supervisor"
	"github.com/sirupsen/logrus"
)

func main() {
	config.LoadDotEnv()
	ll, err := logrus.ParseLevel(config.Config().LogLevel)
	if err != nil {
		logrus.Errorf("unknown log level %s", config.Config().LogLevel)
		ll = logrus.WarnLevel
	}
	logrus.SetLevel(ll)
	supr, err := supervisor.NewSupervisor(config.Config().Supervisor, config.Config().Upstream)
	if err != nil {
		logrus.WithError(err).Fatal("error creating supervisor")
	}
	supr.Start()
}
