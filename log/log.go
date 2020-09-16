package log

import (
	"github.com/sirupsen/logrus"
	"io"
	"os"
)

var Log = logrus.New()

func Init(filename string) error {
	logFile, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0777)
	if err != nil {
		return err
	}
	Log.SetOutput(io.MultiWriter(os.Stdout, logFile))
	Log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	Log.SetReportCaller(true)
	return nil
}
