// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2019 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package logging

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"runtime"
	"time"

	"github.com/kusanagi/kusanagi-sdk-go/transform"
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

var currentLevel = NOTSET

// Mappings between level names and values
var levels = map[string]int{
	"EMERGENCY": EMERGENCY,
	"ALERT":     ALERT,
	"CRITICAL":  CRITICAL,
	"ERROR":     ERROR,
	"WARNING":   WARNING,
	"NOTICE":    NOTICE,
	"INFO":      INFO,
	"DEBUG":     DEBUG,
}

func init() {
	// Remove prefix and flags from default standard logger
	log.SetPrefix("")
	log.SetFlags(0)
}

func getLevelName(level int) string {
	for name, v := range levels {
		if v == level {
			return name
		}
	}
	return ""
}

// Logging date layout
const layout = "2006-01-02T15:04:05.000Z"

func printLog(level int, v ...interface{}) {
	if level > currentLevel {
		return
	}

	prefix := fmt.Sprintf("%s [%s] [SDK]", time.Now().UTC().Format(layout), getLevelName(level))
	log.Println(prefix, fmt.Sprint(v...))
}

func printLogf(level int, format string, v ...interface{}) {
	if level > currentLevel {
		return
	}

	prefix := fmt.Sprintf("%s [%s] [SDK]", time.Now().UTC().Format(layout), getLevelName(level))
	log.Println(prefix, fmt.Sprintf(format, v...))
}

// SetLevel sets current log level.
func SetLevel(level int) {
	currentLevel = level
}

// GetLevel gets current log level.
func GetLevel() int {
	return currentLevel
}

// SetOutput changes logging output.
func SetOutput(w io.Writer) {
	log.SetOutput(w)
}

// Disable disables logging.
func Disable() {
	log.SetOutput(ioutil.Discard)
}

// Enable enables logging.
func Enable() {
	log.SetOutput(os.Stdout)
}

// Emergency logs an emergency message.
func Emergency(v ...interface{}) {
	printLog(EMERGENCY, v...)
}

// Emergencyf logs a emergency message with format.
func Emergencyf(format string, v ...interface{}) {
	printLogf(EMERGENCY, format, v...)
}

// Alert logs an alert message.
func Alert(v ...interface{}) {
	printLog(ALERT, v...)
}

// Alertf logs a alert message with format.
func Alertf(format string, v ...interface{}) {
	printLogf(ALERT, format, v...)
}

// Critical logs a critical message.
func Critical(v ...interface{}) {
	printLog(CRITICAL, v...)
}

// Criticalf logs a critical message with format.
func Criticalf(format string, v ...interface{}) {
	printLogf(CRITICAL, format, v...)
}

// Error logs an error message.
func Error(v ...interface{}) {
	printLog(ERROR, v...)
}

// Errorf logs an error message with format.
func Errorf(format string, v ...interface{}) {
	printLogf(ERROR, format, v...)
}

// Warn logs a warning message.
func Warn(v ...interface{}) {
	printLog(WARNING, v...)
}

// Warnf logs a warning message with format.
func Warnf(format string, v ...interface{}) {
	printLogf(WARNING, format, v...)
}

// Notice logs a notice message.
func Notice(v ...interface{}) {
	printLog(NOTICE, v...)
}

// Noticef logs a notice message with format.
func Noticef(format string, v ...interface{}) {
	printLogf(NOTICE, format, v...)
}

// Info logs an info message.
func Info(v ...interface{}) {
	printLog(INFO, v...)
}

// Infof logs an info message with format.
func Infof(format string, v ...interface{}) {
	printLogf(INFO, format, v...)
}

// Debug logs a debug message.
func Debug(v ...interface{}) {
	printLog(DEBUG, v...)
}

// Debugf logs a debug message with format.
func Debugf(format string, v ...interface{}) {
	printLogf(DEBUG, format, v...)
}

// Json logs a JSON string for debugging.
func Json(v []byte) {
	if v == nil {
		return
	}

	var out bytes.Buffer
	json.Indent(&out, v, "", "  ")
	printLog(DEBUG, out.String())
}

// DebugValue writes a string representation of a value value to the logs.
// See: https://github.com/kusanagi/kusanagi-spec-sdk/blob/master/README.md#34-string-representation
func DebugValue(level int, v interface{}) error {
	var s string

	t := reflect.TypeOf(v)
	if t == nil {
		s = "NULL"
	} else {
		kind := t.Kind()
		switch kind {
		case reflect.Bool:
			if v.(bool) {
				s = "TRUE"
			} else {
				s = "FALSE"
			}
		case reflect.Slice, reflect.Map:
			json, err := transform.Serialize(v, true)
			if err != nil {
				return err
			}
			s = string(json)
		case reflect.Func:
			s = fmt.Sprintf("[function %v]", runtime.FuncForPC(reflect.ValueOf(v).Pointer()).Name())
		default:
			s = fmt.Sprintf("%v", v)
		}
	}
	// Limit the maximum log entry length
	if max := 100000; len(s) > max {
		s = s[:max]
	}
	printLog(level, s)
	return nil
}
