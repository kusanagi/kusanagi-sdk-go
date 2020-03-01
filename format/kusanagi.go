// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2020 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package format

import "fmt"

// ServiceString formats a service name and version into a single string.
func ServiceString(name, version, address string) string {
	v := fmt.Sprintf(`"%v" (%v)`, name, version)
	if address != "" {
		v = fmt.Sprintf(`["%v"] %v`, address, v)
	}
	return v
}
