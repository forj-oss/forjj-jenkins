package reportlogs

import (
	"log"

	"github.com/forj-oss/forjj-modules/trace"
)


type reportLogs struct {
	logsFunc map[string]func(string, ...interface{})

	reportLogFunc func(string, ...interface{})
	reportErrFunc func(string, ...interface{})
	logFunc       func(string, ...interface{})
	bErr          bool
}

func (r *reportLogs) setLogsFunc(logsFunc map[string]func(string, ...interface{})) {
	if r == nil || logsFunc == nil {
		return
	}

	r.logsFunc = logsFunc
	if f, found := logsFunc["reportLog"]; found {
		r.reportLogFunc = f
	} else {
		r.reportLogFunc = func(format string, parameters ...interface{}) {
			log.Printf(format, parameters...)
		}
	}
	if f, found := logsFunc["reportError"]; found {
		r.reportErrFunc = f
	} else {
		r.reportErrFunc = func(format string, parameters ...interface{}) {
			gotrace.Error(format, parameters...)
		}
	}
	if f, found := logsFunc["pluginLog"]; found {
		r.logFunc = f
	} else {
		r.logFunc = func(format string, parameters ...interface{}) {
			log.Printf(format, parameters...)
		}
	}

}

// reportLog call the log function registered
func (r *reportLogs) reportLog(format string, parameters ...interface{}) {
	if r == nil || r.logsFunc == nil {
		return
	}

	r.reportLogFunc(format, parameters...)
}

// reportError call the log function registered
func (r *reportLogs) reportError(format string, parameters ...interface{}) {
	if r == nil || r.logsFunc == nil {
		return
	}

	r.reportErrFunc(format, parameters...)
	r.bErr = true
}

// reportError call the log function registered
func (r *reportLogs) log(format string, parameters ...interface{}) {
	if r == nil || r.logsFunc == nil {
		return
	}

	r.logFunc(format, parameters...)
}

func (r *reportLogs) hasReportedErrors() bool {
	if r == nil || r.logsFunc == nil {
		gotrace.Error("reportLogs is nil or has no valid log handler")
		return true
	}
	bErr := r.bErr
	r.bErr = false
	return bErr
}

