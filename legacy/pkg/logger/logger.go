package logger

import (
	"log"
	"os"

	"github.com/sirupsen/logrus"
)

var Logger = logrus.New()

func InitLogger() {
	// Include file name and line number in log output
	log.SetFlags(log.LstdFlags | log.Llongfile)

	// logrus
	Logger.SetReportCaller(true) // line of code
	Logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05", // custom format
	})
	Logger.SetLevel(logrus.InfoLevel)
	Logger.Out = os.Stdout
	log.SetOutput(Logger.Writer())
}
