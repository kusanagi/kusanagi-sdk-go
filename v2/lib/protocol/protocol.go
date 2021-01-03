// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2020 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package protocol

import (
	"fmt"
	"regexp"
	"strings"
)

// Regexp to parse the addresses to be used as IPC names.
var ipcRegexp = regexp.MustCompile("[^a-zA-Z0-9]{1,}")

// IPC creates an IPC connection string.
func IPC(args ...string) string {
	name := ipcRegexp.ReplaceAllString(strings.Join(args, "-"), "-")
	return fmt.Sprintf("ipc://@kusanagi-%s", name)
}

// SocketAddress creates a ZMQ socket address.
func SocketAddress(address string, tcp bool) string {
	// Check if TCP must be used
	if tcp {
		return fmt.Sprintf("tcp://%s", address)
	}
	// Otherwise use IPC
	return IPC(address)
}
