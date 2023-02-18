// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2022 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package log

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"runtime"
	"time"

	"github.com/kusanagi/kusanagi-sdk-go/v5/lib/json"
)

// These flags define the error levels following RFC5425.
// See https://tools.ietf.org/html/rfc5424/.
const (
	NOTSET = iota - 1
	EMERGENCY
	ALERT
	CRITICAL
	ERROR
	WARNING
	NOTICE
	INFO
	DEBUG
)

// Remove prefix and flags from default standard logger.
func init() {
	log.SetPrefix("")
	log.SetFlags(0)
}

// SetOutput changes the logging output.
func SetOutput(w io.Writer) {
	log.SetOutput(w)
}

// Disable logging.
func Disable() {
	log.SetOutput(ioutil.Discard)
}

// Enable logging.
func Enable() {
	log.SetOutput(os.Stdout)
}

// The level currently selected.
var currentLevel = ERROR

// SetLevel changes the current log level.
func SetLevel(level int) {
	currentLevel = level
}

// GetLevel returns the current log level.
func GetLevel() int {
	return currentLevel
}

var levels = map[int]string{
	EMERGENCY: "EMERGENCY",
	ALERT:     "ALERT",
	CRITICAL:  "CRITICAL",
	ERROR:     "ERROR",
	WARNING:   "WARNING",
	NOTICE:    "NOTICE",
	INFO:      "INFO",
	DEBUG:     "DEBUG",
}

func getLogPrefix(level int) string {
	timestamp := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")

	return fmt.Sprintf("%s [%s] [SDK]", timestamp, levels[level])
}

// Log writes a log message.
func Log(level int, v ...interface{}) {
	if level <= currentLevel {
		log.Println(getLogPrefix(level), fmt.Sprint(v...))
	}
}

// Logf writes a log message for a level with format.
func Logf(level int, format string, v ...interface{}) {
	if level <= currentLevel {
		log.Println(getLogPrefix(level), fmt.Sprintf(format, v...))
	}
}

// Emergency logs an emergency message.
func Emergency(v ...interface{}) {
	Log(EMERGENCY, v...)
}

// Emergencyf logs a emergency message with format.
func Emergencyf(format string, v ...interface{}) {
	Logf(EMERGENCY, format, v...)
}

// Alert logs an alert message.
func Alert(v ...interface{}) {
	Log(ALERT, v...)
}

// Alertf logs a alert message with format.
func Alertf(format string, v ...interface{}) {
	Logf(ALERT, format, v...)
}

// Critical logs a critical message.
func Critical(v ...interface{}) {
	Log(CRITICAL, v...)
}

// Criticalf logs a critical message with format.
func Criticalf(format string, v ...interface{}) {
	Logf(CRITICAL, format, v...)
}

// Error logs an error message.
func Error(v ...interface{}) {
	Log(ERROR, v...)
}

// Errorf logs an error message with format.
func Errorf(format string, v ...interface{}) {
	Logf(ERROR, format, v...)
}

// Warning logs a warning message.
func Warning(v ...interface{}) {
	Log(WARNING, v...)
}

// Warningf logs a warning message with format.
func Warningf(format string, v ...interface{}) {
	Logf(WARNING, format, v...)
}

// Notice logs a notice message.
func Notice(v ...interface{}) {
	Log(NOTICE, v...)
}

// Noticef logs a notice message with format.
func Noticef(format string, v ...interface{}) {
	Logf(NOTICE, format, v...)
}

// Info logs an info message.
func Info(v ...interface{}) {
	Log(INFO, v...)
}

// Infof logs an info message with format.
func Infof(format string, v ...interface{}) {
	Logf(INFO, format, v...)
}

// Debug logs a debug message.
func Debug(v ...interface{}) {
	Log(DEBUG, v...)
}

// Debugf logs a debug message with format.
func Debugf(format string, v ...interface{}) {
	Logf(DEBUG, format, v...)
}

// NewRequestLogger creates a new logger wuth request ID support.
func NewRequestLogger(rid string) RequestLogger {
	if rid == "" {
		rid = "-"
	}

	return RequestLogger{rid, fmt.Sprintf(" |%s|", rid)}
}

// RequestLogger is a logger with request ID support.
// The request ID is added to every log message written using this logger.
type RequestLogger struct {
	rid    string
	suffix string
}

// RID returns the request ID.
func (r RequestLogger) RID() string {
	return r.rid
}

// Emergency logs an emergency message.
func (r RequestLogger) Emergency(v ...interface{}) {
	Emergency(append(v, r.suffix)...)
}

// Emergencyf logs a emergency message with format.
func (r RequestLogger) Emergencyf(format string, v ...interface{}) {
	Emergencyf(format+r.suffix, v...)
}

// Alert logs an alert message.
func (r RequestLogger) Alert(v ...interface{}) {
	Alert(append(v, r.suffix)...)
}

// Alertf logs a alert message with format.
func (r RequestLogger) Alertf(format string, v ...interface{}) {
	Alertf(format+r.suffix, v...)
}

// Critical logs a critical message.
func (r RequestLogger) Critical(v ...interface{}) {
	Critical(append(v, r.suffix)...)
}

// Criticalf logs a critical message with format.
func (r RequestLogger) Criticalf(format string, v ...interface{}) {
	Criticalf(format+r.suffix, v...)
}

// Error logs an error message.
func (r RequestLogger) Error(v ...interface{}) {
	Error(append(v, r.suffix)...)
}

// Errorf logs an error message with format.
func (r RequestLogger) Errorf(format string, v ...interface{}) {
	Errorf(format+r.suffix, v...)
}

// Warning logs a warning message.
func (r RequestLogger) Warning(v ...interface{}) {
	Warning(append(v, r.suffix)...)
}

// Warningf logs a warning message with format.
func (r RequestLogger) Warningf(format string, v ...interface{}) {
	Warningf(format+r.suffix, v...)
}

// Notice logs a notice message.
func (r RequestLogger) Notice(v ...interface{}) {
	Notice(append(v, r.suffix)...)
}

// Noticef logs a notice message with format.
func (r RequestLogger) Noticef(format string, v ...interface{}) {
	Noticef(format+r.suffix, v...)
}

// Info logs an info message.
func (r RequestLogger) Info(v ...interface{}) {
	Info(append(v, r.suffix)...)
}

// Infof logs an info message with format.
func (r RequestLogger) Infof(format string, v ...interface{}) {
	Infof(format+r.suffix, v...)
}

// Debug logs a debug message.
func (r RequestLogger) Debug(v ...interface{}) {
	Debug(append(v, r.suffix)...)
}

// Debugf logs a debug message with format.
func (r RequestLogger) Debugf(format string, v ...interface{}) {
	Debugf(format+r.suffix, v...)
}

// Log a message.
func (r RequestLogger) Log(level int, v ...interface{}) {
	Log(level, append(v, r.suffix)...)
}

// ValueToLogString returns a string representation of a value.
func ValueToLogString(value interface{}) (result string, err error) {
	valueType := reflect.TypeOf(value)
	if valueType == nil {
		return "NULL", nil
	}

	switch valueType.Kind() {
	case reflect.Bool:
		if v, _ := value.(bool); v {
			result = "TRUE"
		} else {
			result = "FALSE"
		}
	case reflect.Slice, reflect.Map:
		if result, err = json.Serialize(value, true); err != nil {
			return "", err
		}
	case reflect.Func:
		result = fmt.Sprintf("[function %v]", runtime.FuncForPC(reflect.ValueOf(value).Pointer()).Name())
	default:
		result = fmt.Sprintf("%v", value)
	}

	// Limit the maximum log entry length
	if max := 100000; len(result) > max {
		result = result[:max]
	}

	return result, nil
}
