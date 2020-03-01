// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2020 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package cli

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

// Options stores all the command line option values
var Options *CommandLineOptions

func init() {
	// Don't print usage help on error
	flag.Usage = func() {}
	// Create an object to store command line option values
	Options = NewCommandLineOptions()
}

// Parse parses and validates command line options.
// This function saves the option values into the global Options object.
func Parse() {
	// Parse command line and set option values
	if err := flag.CommandLine.Parse(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// When help is present display help and exit
	if Options.Help {
		Options.PrintUsage()
		os.Exit(0)
	} else if err := Options.Validate(); err != nil {
		// When option value validation fails print error and exit
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type ErrRequiredOption struct {
	Name string
}

func (e ErrRequiredOption) Error() string {
	return fmt.Sprintf("required option: %v", e.Name)
}

type ErrInvalidOptionValue struct {
	Name string
}

func (e ErrInvalidOptionValue) Error() string {
	return fmt.Sprintf("invalid option value: %v", e.Name)
}

// NewCommandLineOptions creates a new CommandLineOptions object and prepares it
// to store all the command line option values.
func NewCommandLineOptions() *CommandLineOptions {
	c := &CommandLineOptions{}
	StringVar(&c.Action, "A", "action", "", "name of action to call when request message is given as JSON through stdin", false)
	StringVar(&c.Component, "c", "component", "", "component type [service|middleware]", true)
	BoolVar(&c.DisableShortNames, "d", "disable-compact-names", false, "use full property names in payloads", false)
	BoolVar(&c.Debug, "D", "debug", false, "enable debugging", false)
	BoolVar(&c.Help, "h", "help", false, "print help", false)
	IntVar(&c.LogLevel, "L", "log-level", 0, "enable logging using numeric syslog severity value [0-7]", false)
	StringVar(&c.Name, "n", "name", "", "component name", true)
	StringVar(&c.FrameworkVersion, "p", "framework-version", "", "KUSANAGI framework version", true)
	StringVar(&c.Socket, "s", "socket", "", "IPC socket name", false)
	IntVar(&c.TCP, "t", "tcp", 0, "TCP port to use when IPC socket is not used", false)
	IntVar(&c.Timeout, "T", "timeout", 30000, "process execution timeout per request in milliseconds", false)
	StringVar(&c.Version, "v", "version", "", "component version", true)
	KeyValueVar(&c.Var, "V", "var", "", "component variables", false)
	return c
}

// CommandLineOptions defines the suported CLI options
type CommandLineOptions struct {
	Action            string
	Component         string
	Debug             bool
	DisableShortNames bool
	FrameworkVersion  string
	Help              bool
	LogLevel          int
	Name              string
	Port              int
	Socket            string
	TCP               int
	Timeout           int
	Var               KeyValueFlag
	Version           string
}

// PrintUsage prints usage and options to stdout
func (c CommandLineOptions) PrintUsage() {
	fmt.Fprintf(os.Stdout, "usage: %s OPTION...\n\n", os.Args[0])
	fmt.Fprintf(os.Stdout, "options:\n")
	fmt.Fprintf(os.Stdout, c.GetHelp())
}

// GetHelp gets a string with the command line options help
func (c CommandLineOptions) GetHelp() string {
	line := ""
	buf := &bytes.Buffer{}
	w := tabwriter.NewWriter(buf, 0, 0, 1, ' ', 0)
	for _, f := range flags {
		if !f.IsBoolean {
			t := strings.Replace(strings.ToUpper(f.Name), "-", "_", -1)
			line = fmt.Sprintf("  -%s %s, --%s %s\t\t%s", f.ShortName, t, f.Name, t, f.Usage)
		} else {
			line = fmt.Sprintf("  -%s, --%s\t\t%s", f.ShortName, f.Name, f.Usage)
		}

		if f.Default != "" {
			line = fmt.Sprintf("%v (default: %v)", line, f.Default)
		} else if f.Required {
			line = fmt.Sprintf("%v (required)", line)
		}

		fmt.Fprintln(w, line)
	}
	w.Flush()
	return buf.String()
}

// Validate validates command line option values
func (c CommandLineOptions) Validate() error {
	if c.Component == "" {
		return ErrRequiredOption{"component"}
	} else if c.Component != "service" && c.Component != "middleware" {
		return ErrInvalidOptionValue{"component"}
	}

	if c.Name == "" {
		return ErrRequiredOption{"name"}
	}

	if c.FrameworkVersion == "" {
		return ErrRequiredOption{"framework-version"}
	}

	if c.Version == "" {
		return ErrRequiredOption{"version"}
	}

	return nil
}
