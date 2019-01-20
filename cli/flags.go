// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2019 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package cli

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
)

var flags = []*Flag{}

type Flag struct {
	Name,
	ShortName,
	Usage,
	Default string
	Required,
	IsBoolean bool
}

type KeyValueFlag struct {
	Values map[string]string
}

func (k *KeyValueFlag) String() string {
	var values []string
	for name, v := range k.Values {
		values = append(values, fmt.Sprintf("%s=%s", name, v))
	}
	// Return all the key values in a single comma separated string.
	// This is done to comply with the "flag.Value" interface.
	return strings.Join(values, ",")
}

func (k *KeyValueFlag) Set(value string) error {
	if k.Values == nil {
		k.Values = make(map[string]string)
	}

	values := strings.SplitN(value, "=", 2)
	if len(values) != 2 {
		return fmt.Errorf("invalid value: %v", value)
	}

	k.Values[values[0]] = values[1]
	return nil
}

func (k *KeyValueFlag) GetUint32(name string) (uint32, bool) {
	if s, ok := k.Values[name]; ok {
		if v, err := strconv.ParseUint(s, 10, 32); err == nil {
			return uint32(v), true
		}
	}
	return 0, false
}

func KeyValueVar(p *KeyValueFlag, short, name, value, usage string, required bool) {
	flag.Var(p, short, usage)
	flag.Var(p, name, usage)
	flags = append(flags, &Flag{name, short, usage, value, required, false})
}

func StringVar(p *string, short, name, value, usage string, required bool) {
	flag.StringVar(p, short, value, usage)
	flag.StringVar(p, name, value, usage)
	flags = append(flags, &Flag{name, short, usage, value, required, false})
}

func BoolVar(p *bool, short, name string, value bool, usage string, required bool) {
	flag.BoolVar(p, short, value, usage)
	flag.BoolVar(p, name, value, usage)
	flags = append(flags, &Flag{name, short, usage, "", required, true})
}

func IntVar(p *int, short, name string, value int, usage string, required bool) {
	flag.IntVar(p, short, value, usage)
	flag.IntVar(p, name, value, usage)

	v := ""
	if value != 0 {
		v = fmt.Sprintf("%v", value)
	}

	flags = append(flags, &Flag{name, short, usage, v, required, false})
}
