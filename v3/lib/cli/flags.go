// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2021 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
)

// List of CLI options.
var options = []option{}

// An option represents a CLI input option.
type option struct {
	name      string
	shortName string
	usage     string
	preset    string
	required  bool
	isBoolean bool
}

func (o option) String() string {
	line := []string{}
	if o.isBoolean {
		line = append(line, fmt.Sprintf("  -%s, --%s\t\t%s", o.shortName, o.name, o.usage))
	} else {
		// When the option is not boolean add a placeholder after the option name.
		// The placeholder is the option name in uppercase.
		v := strings.Replace(strings.ToUpper(o.name), "-", "_", -1)
		line = append(line, fmt.Sprintf("  -%s %s, --%s %s\t\t%s", o.shortName, v, o.name, v, o.usage))
	}

	if o.preset != "" {
		line = append(line, fmt.Sprintf("(default: %v)", o.preset))
	} else if o.required {
		line = append(line, "(required)")
	}
	return strings.Join(line, " ")
}

// PrintHelp prints the command line options help.
func PrintHelp(out io.Writer) {
	fmt.Fprintf(out, "usage: %s OPTION...\n\n", os.Args[0])
	fmt.Fprintf(out, "options:\n")
	w := tabwriter.NewWriter(out, 0, 0, 1, ' ', 0)
	for _, option := range options {
		fmt.Fprintln(w, option)
	}
	w.Flush()
}

type keyValue map[string]string

func (k keyValue) String() string {
	// Join the values into their CLI original values
	values := []string{}
	for name, value := range k {
		values = append(values, fmt.Sprintf("%s=%s", name, value))
	}
	// Return all the key values in a single comma separated string.
	// This is done to comply with the "flag.Value" interface.
	return strings.Join(values, ",")
}

func (k keyValue) Set(value string) error {
	// Split the CLI input value into the key and the value
	if values := strings.SplitN(value, "=", 2); len(values) == 2 {
		k[values[0]] = values[1]
	} else {
		return fmt.Errorf("invalid value: %s", value)
	}
	return nil
}

func stringOption(short, name, usage, preset string, required bool) *string {
	flag.StringVar(&preset, short, preset, usage)
	flag.StringVar(&preset, name, preset, usage)
	options = append(options, option{name, short, usage, preset, required, false})
	return &preset
}

func boolOption(short, name, usage string, preset bool, required bool) *bool {
	flag.BoolVar(&preset, short, preset, usage)
	flag.BoolVar(&preset, name, preset, usage)
	options = append(options, option{name, short, usage, "", required, true})
	return &preset
}

func uintOption(short, name, usage string, preset uint, required bool) *uint {
	flag.UintVar(&preset, short, preset, usage)
	flag.UintVar(&preset, name, preset, usage)
	options = append(options, option{name, short, usage, fmt.Sprintf("%d", preset), required, false})
	return &preset
}

func intOption(short, name, usage string, preset int, required bool) *int {
	flag.IntVar(&preset, short, preset, usage)
	flag.IntVar(&preset, name, preset, usage)
	options = append(options, option{name, short, usage, fmt.Sprintf("%d", preset), required, false})
	return &preset
}

func keyValueOption(short, name, usage string, required bool) keyValue {
	v := keyValue{}
	flag.Var(v, short, usage)
	flag.Var(v, name, usage)
	options = append(options, option{name, short, usage, "", required, false})
	return v
}
