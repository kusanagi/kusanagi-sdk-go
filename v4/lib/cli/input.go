// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2022 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/kusanagi/kusanagi-sdk-go/v4/lib/log"
)

// List of CLI options.
var component = stringOption(
	"c", "component",
	"Component type [service|middleware]",
	"",
	true,
)
var debug = boolOption(
	"D", "debug",
	"Enable debugging",
	false,
	false,
)
var help = boolOption(
	"h", "help",
	"Print help",
	false,
	false,
)
var socket = stringOption(
	"i", "ipc",
	"IPC socket name",
	"",
	false,
)
var logLevel = uintOption(
	"L", "log-level",
	"Enable logging using a numeric syslog severity value [0-7]",
	0,
	false,
)
var name = stringOption(
	"n", "name",
	"Component name",
	"",
	true,
)
var frameworkVersion = stringOption(
	"p", "framework-version",
	"KUSANAGI framework version",
	"",
	true,
)
var tcp = uintOption(
	"t", "tcp",
	"TCP port to use when IPC socket is not used",
	0,
	false,
)
var timeout = intOption(
	"T", "timeout",
	"Process execution timeout per request in milliseconds",
	30000,
	false,
)
var version = stringOption(
	"v", "version",
	"Component version",
	"",
	true,
)
var vars = keyValueOption(
	"V", "var",
	"Component variables",
	false,
)

func init() {
	// Don't print usage help on error
	flag.Usage = func() {}
}

func newErrRequired(name string) error {
	return fmt.Errorf(`required option: "%s"`, name)
}

func newErrInvalid(name string) error {
	return fmt.Errorf(`invalid value for option: "%s"`, name)
}

// Parse processes and validates command line options.
//
// The result is an input object that allows access to the CLI option values.
func Parse() (Input, error) {
	input := Input{}

	// Parse command line and set option values
	if err := flag.CommandLine.Parse(os.Args[1:]); err != nil {
		return input, err
	}

	// Validate the option values when no help must be displayed
	if *help {
		PrintHelp(os.Stderr)
		os.Exit(0)
	} else {
		if component == nil || *component == "" {
			return input, newErrRequired("component")
		} else if v := *component; v != "service" && v != "middleware" {
			return input, newErrInvalid("component")
		} else if name == nil || *name == "" {
			return input, newErrRequired("name")
		} else if frameworkVersion == nil || *frameworkVersion == "" {
			return input, newErrRequired("framework-version")
		} else if version == nil || *version == "" {
			return input, newErrRequired("version")
		}
	}

	// Get the name of the running executable to use as input path
	path, err := os.Executable()
	if err != nil {
		log.Warningf("failed to get executable path, using run arguments: %v", err)
		path = os.Args[0]
	}

	input.path = path
	return input, nil
}

// Input contains the CLI input values
type Input struct {
	path string
}

// GetPath returns the path to the file being executed.
//
// The path includes the file name.
func (i Input) GetPath() string {
	return i.path
}

// MustDisplayHelp checks if the CLI help must be displayed.
func (i Input) MustDisplayHelp() bool {
	return help != nil && *help
}

// GetComponent returns the component type.
func (i Input) GetComponent() string {
	if component == nil {
		return ""
	}
	return *component
}

// GetName returns the component name.
func (i Input) GetName() string {
	if name == nil {
		return ""
	}
	return *name
}

// GetVersion returns the component version.
func (i Input) GetVersion() string {
	if version == nil {
		return ""
	}
	return *version
}

// GetComponentTitle returns the component name and version.
func (i Input) GetComponentTitle() string {
	return fmt.Sprintf(`"%s" (%s)`, i.GetName(), i.GetVersion())
}

// GetFrameworkVersion returns the KUSANAGI framework version.
func (i Input) GetFrameworkVersion() string {
	if frameworkVersion == nil {
		return ""
	}
	return *frameworkVersion
}

// GetTCP returns the port to use for TCP connections.
func (i Input) GetTCP() uint {
	if tcp == nil {
		return 0
	}
	return *tcp
}

// IsTCPEnabled checks if TCP connections should be used instead of IPC.
func (i Input) IsTCPEnabled() bool {
	return i.GetTCP() != 0
}

// GetSocket returns the ZMQ socket name.
func (i Input) GetSocket() string {
	if socket == nil || i.IsTCPEnabled() {
		return ""
	}
	return *socket
}

// GetTimeout returns the process execution timeout in milliseconds.
func (i Input) GetTimeout() int {
	if timeout == nil {
		return 0
	}
	return *timeout
}

// IsDebugEnabled checks if debug is enabled.
func (i Input) IsDebugEnabled() bool {
	return debug != nil && *debug
}

// HasVariable checks if an engine variable is defined.
//
// name: The name of the variable.
func (i Input) HasVariable(name string) bool {
	_, exists := vars[name]
	return exists
}

// GetVariable returns the value for an engine variable.
//
// name: The name of the variable.
func (i Input) GetVariable(name string) string {
	return vars[name]
}

// GetVariables returns all the engine variables.
func (i Input) GetVariables() map[string]string {
	variables := make(map[string]string)
	if vars != nil {
		for name, value := range vars {
			variables[name] = value
		}
	}
	return variables
}

// HasLogging checks if logging is enabled.
func (i Input) HasLogging() bool {
	return logLevel != nil
}

// GetLogLevel returns the log level.
//
// The INFO level is returned when no log level is defined.
func (i Input) GetLogLevel() int {
	if logLevel == nil {
		return log.INFO
	}
	return int(*logLevel)
}
