// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2020 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package format

import "time"

// Format to be used for dates (nanoseconds are required for better accuracy)
const dateLayout = "2006-01-02T15:04:05.000+00:00"

// TimeToString converts a time to a string (UTC ISO8601).
func TimeToString(t time.Time) string {
	return t.Format(dateLayout)
}

// TimeFromString converts a date string (UTC ISO8601) to time.
func TimeFromString(s string) (time.Time, error) {
	return time.Parse(dateLayout, s)
}
