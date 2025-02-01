package utils

import (
	"os"

	log "github.com/sirupsen/logrus"
)

var Log = log.New()

func InitializeLogger() {
	Log.SetFormatter(&log.JSONFormatter{})
	Log.SetOutput(os.Stdout)
	Log.SetLevel(log.InfoLevel)
}
