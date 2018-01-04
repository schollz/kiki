package logging

import (
	"fmt"

	seelog "github.com/cihub/seelog"
)

var (
	Verbose      bool = false
	Log          seelog.LoggerInterface
	LogDirectory string = "log"
	LogLevel     string = "trace"
)

func initLogging() {
	Log = seelog.Disabled

	// https://github.com/cihub/seelog/wiki/Log-levels
	appConfig := `
<seelog minlevel="` + LogLevel + `">
    <outputs formatid="stdout">
        <filter levels="info,debug,trace,critical,error,warn">
            <console formatid="stdout"/>
        </filter>
    </outputs>
    <formats>
        <format id="stdout"   format="%Date %Time [%LEVEL] %RelFile %FuncShort:%Line %Msg %n" />
    </formats>
</seelog>
`
	// <format id="common"   format="%Date %Time [%LEVEL] %File %FuncShort:%Line %Msg %n" />

	logger, err := seelog.LoggerFromConfigAsBytes([]byte(appConfig))
	if err != nil {
		fmt.Println(err)
		return
	}
	Log = logger
}

func init() {
	initLogging()
}

func Debug(t bool) {
	if t {
		LogLevel = "debug"
	} else {
		LogLevel = "warn"
	}
	initLogging()
}

// package logging
//
// import (
// 	"fmt"
// 	"github.com/sirupsen/logrus"
// 	"os"
// )
//
// var Log = logrus.New()
//
// type KikiFormatter struct{}
//
// func (self *KikiFormatter) Format(entry *logrus.Entry) ([]byte, error) {
// 	fmt.Println(entry)
// 	msg := fmt.Sprintf("%v [%v] %v %v", entry.Time, entry.Level, entry.Message, entry.Data)
// 	// fmt.Println(entry.String())
// 	return []byte(msg + "\n"), nil
// }
//
// func Setup() {
// 	// Log as JSON instead of the default ASCII formatter.
// 	// log.SetFormatter(&log.JSONFormatter{})
// 	logrus.SetFormatter(new(KikiFormatter))
// 	logrus.Info("TEST")
//
// 	// Output to stdout instead of the default stderr
// 	// Can be any io.Writer, see below for File example
// 	Log.Out = os.Stdout
//
// 	// Only log the warning severity or above.
// 	Log.SetLevel(logrus.DebugLevel)
// }
//
// // Debug will switch the verbosity of the database.
// func Debug(t bool) {
// 	if t {
// 		Log.Level = logrus.DebugLevel
// 	} else {
// 		Log.Level = logrus.WarnLevel
// 	}
// }
