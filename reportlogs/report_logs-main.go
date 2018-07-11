package reportlogs

// Log systems to report internally or externally.
var modLog reportLogs

// SetLogsFunc initialize logs functions.
// We expect to have "log" and "err" functions
// and will be registered.
//
// Those function are used to report to end users by forjj
// Detailled log are still stored in the plugin log output.
func SetLogsFunc(logsFunc map[string]func(string, ...interface{})) {
	modLog.setLogsFunc(logsFunc)
}

// Reportf call the log function registered
func Reportf(format string, parameters ...interface{}) {
	modLog.reportLog(format, parameters...)
}

// Errorf call the log function registered
func Errorf(format string, parameters ...interface{}) {
	modLog.reportError(format, parameters...)
}

// Printf to native log system or registered log mechanism
func Printf(format string, parameters ...interface{}) {
	modLog.log(format, parameters...)
}

// HasReportedErrors return if since last call to this function
// errors were reported. Return true in that case. false otherwise.
func HasReportedErrors() bool {
	return modLog.hasReportedErrors()
}