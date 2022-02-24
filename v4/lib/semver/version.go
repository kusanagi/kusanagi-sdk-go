// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2022 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package semver

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/kusanagi/kusanagi-sdk-go/v4/lib/log"
)

// Regexp to check for invalid chars in version patterns
var reInvalidVersionChars = regexp.MustCompile(`[^a-zA-Z0-9*.,_-]`)

// Regexp to match duplicated wildcards
var reWildcards = regexp.MustCompile(`\*+`)

// Regexp to match version dot separators
var reVersionDots = regexp.MustCompile(`([^*])\.`)

// Regexp to match wildcards, with the exception of the suffix ones
var reVersionWildcards = regexp.MustCompile(`\*+([^$])`)

func compareEmpty(part1, part2 string) bool {
	// The one that DO NOT have more parts is greater
	return (part1 != part2) && (part2 == "")
}

func compareSubParts(sub1, sub2 string) bool {
	if sub1 == sub2 {
		return false
	}

	// Check if any sub part is an integer
	isInteger := [2]bool{false, false}
	parts := [2]string{sub1, sub2}
	for i, sub := range parts {
		_, err := strconv.Atoi(sub)
		if err == nil {
			isInteger[i] = true
		}
	}

	// When one is an integer and the other a string, the integer
	// is always higher than the string one.
	if isInteger[0] != isInteger[1] {
		// Check if the first sub part is an integer, and if so it means sub2 is lower than sub1
		return isInteger[1]
	}

	// Both sub parts are of the same type
	return sub1 < sub2
}

func zipVersionParts(items ...[]string) [][]string {
	// Get length for each slice
	maxLength := 0
	lengths := make([]int, len(items))
	for i, item := range items {
		lengths[i] = len(item)
		if lengths[i] > maxLength {
			maxLength = lengths[i]
		}
	}

	// Iterate all elements
	var zip [][]string
	for i := 0; i < maxLength; i++ {
		current := []string{}
		// Iterate each of the slices once per slice for current index
		for pos, item := range items {
			if i < lengths[pos] {
				current = append(current, item[i])
			} else {
				// Add empty when no more elements are available in current slice
				current = append(current, "")
			}
		}
		zip = append(zip, current)
	}
	return zip
}

func compareVersions(versions []string) func(i, j int) bool {
	// Return the sorting function
	return func(i, j int) bool {
		ver1 := versions[i]
		ver2 := versions[j]
		if ver1 == ver2 {
			return false
		}

		// Split versions into parts
		vp := [][]string{strings.Split(ver1, "."), strings.Split(ver2, ".")}

		// Iterate over part tuples
		for _, parts := range zipVersionParts(vp...) {
			// One of the parts is empty
			if (parts[0] == "") || (parts[1] == "") {
				return compareEmpty(parts[0], parts[1])
			}

			// Split parts into sub parts
			sp := [][]string{strings.Split(parts[0], "-"), strings.Split(parts[1], "-")}

			// Iterate over sub part tuples
			for _, subParts := range zipVersionParts(sp...) {
				// One of the sub parts is empty
				if (subParts[0] == "") || (subParts[1] == "") {
					return compareEmpty(subParts[0], subParts[1])
				}

				// Both sub parts have a value
				result := compareSubParts(subParts[0], subParts[1])
				if result {
					// Sub parts are not equal
					return result
				}
			}
		}
		return false
	}
}

// New creates a new semantic version.
func New(pattern string) Version {
	// Remove duplicated '*' from the version pattern
	v := Version{value: reWildcards.ReplaceAllString(pattern, "*")}

	if strings.Contains(v.value, "*") {
		// Create an expression to use for version pattern comparison (${1} is the extra matched char)
		expr := reVersionWildcards.ReplaceAllString(v.value, `[^*.]+${1}`)
		// Escape dots to work with the regular expression (${1} adds the matched prefix char)
		expr = reVersionDots.ReplaceAllString(expr, `${1}\.`)

		// If there is a final wildcard left replace it with an expression to match any characters after the last dot
		lc := len(expr) - 1
		if string(expr[lc]) == "*" {
			expr = expr[:lc] + ".*"
		}

		// Create a pattern to be used for comparison
		pattern, err := regexp.Compile(fmt.Sprintf("^%v$", expr))
		if err != nil {
			log.Errorf(`failed to parse version pattern "%s": %v`, v.value, err)
		} else {
			v.pattern = pattern
		}
	}
	return v
}

// Version is used to resolve component versions.
type Version struct {
	pattern *regexp.Regexp
	value   string
}

// AllowWildcards checks if version patterns with wildcards are supported by the version.
func (v Version) AllowWildcards() bool {
	return v.pattern != nil
}

// Match checks if a version matches the current version pattern.
func (v Version) Match(version string) bool {
	// Check that the version pattern is valid
	if reInvalidVersionChars.MatchString(v.value) {
		return false
	}

	if v.pattern != nil {
		return v.pattern.MatchString(version)
	}
	return v.value == version
}

// Resolve resolves to the highest compatible version.
// An empty string is returned when the no version is resolved.
func (v Version) Resolve(versions []string) string {
	// Filter out versions that don't match current version pattern
	var compatible []string
	for _, version := range versions {
		// Append versions that match current version
		if v.Match(version) {
			compatible = append(compatible, version)
		}
	}

	if len(compatible) == 0 {
		return ""
	}

	// Sort the compatible versions and return the higher one
	sort.Slice(compatible, compareVersions(compatible))
	return compatible[len(compatible)-1]
}
