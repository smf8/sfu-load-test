package log

import (
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

const Level = "debug"

func SetupLogger() {
	logLevel, err := logrus.ParseLevel(Level)
	if err != nil {
		logLevel = logrus.ErrorLevel
	}

	logrus.SetLevel(logLevel)
	logrus.SetOutput(os.Stdout)

	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339,
	})
}
