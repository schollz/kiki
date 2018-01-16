package logging

import (
	"fmt"
	"os"

	seelog "github.com/cihub/seelog"
)

// https://github.com/cihub/seelog/wiki/Custom-formatters
func pidLogFormatter(params string) seelog.FormatterFunc {
	return func(message string, level seelog.LogLevel, context seelog.LogContextInterface) interface{} {
		var pid = os.Getpid()
		return fmt.Sprintf("%v", pid)
	}
}

func init() {
	seelog.RegisterCustomFormatter("pidLogFormatter", pidLogFormatter)
}
