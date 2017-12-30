package logging

import (
	"os"

	"github.com/sirupsen/logrus"
)

var Log = logrus.New()

func Setup() {
	// Log as JSON instead of the default ASCII formatter.
	// log.SetFormatter(&log.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	Log.Out = os.Stdout

	// Only log the warning severity or above.
	Log.SetLevel(logrus.InfoLevel)
}

// Debug will switch the verbosity of the database.
func Debug(t bool) {
	if t {
		Log.Level = logrus.DebugLevel
	} else {
		Log.Level = logrus.WarnLevel
	}
}
