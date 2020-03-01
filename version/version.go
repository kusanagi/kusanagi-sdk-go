// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2020 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package version

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// SDK version
const version = "2.0.0"

// Get the KUSANAGI SDK version.
func Get() string {
	return version
}

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
	lenghts := make([]int, len(items))
	for i, item := range items {
		lenghts[i] = len(item)
		if lenghts[i] > maxLength {
			maxLength = lenghts[i]
		}
	}

	// Iterate all elements
	var zip [][]string
	for i := 0; i < maxLength; i++ {
		current := []string{}
		// Iterate each of the slices once per slice for current index
		for pos, item := range items {
			if i < lenghts[pos] {
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

// ErrVersionNotFound is used when a version is not found for a version pattern.
type ErrVersionNotFound struct {
	Version string
}

func (e *ErrVersionNotFound) Error() string {
	return fmt.Sprintf("Service version not found for pattern: \"%s\"", e.Version)
}

// ErrInvalidVersion is used when a version contains invalid characters.
type ErrInvalidVersion struct {
	Version string
}

func (e *ErrInvalidVersion) Error() string {
	return fmt.Sprintf("Invalid version pattern: \"%s\"", e.Version)
}

// New creates a new version for a pattern.
func New(version string) (*Version, error) {
	v := &Version{}
	if err := v.SetValue(version); err != nil {
		return nil, err
	}
	return v, nil
}

// Version is used to resolve component versions.
type Version struct {
	pattern *regexp.Regexp
	value   string
}

// SetValue sets the version value.
func (v *Version) SetValue(value string) error {
	// Validate that version does not contain invalid characters
	if reInvalidVersionChars.MatchString(value) {
		return &ErrInvalidVersion{Version: value}
	}

	// When version contains wildcards remove duplicated ones and create a pattern for comparison
	if strings.Contains(value, "*") {
		var err error

		// Remove duplicated '*' from version
		value = reWildcards.ReplaceAllString(value, "*")

		// Create an expression to use for version pattern comparison (${1} is the extra matched char)
		expr := reVersionWildcards.ReplaceAllString(value, `[^*.]+${1}`)
		// Escape dots to work with the regular expression (${1} adds the matched prefix char)
		expr = reVersionDots.ReplaceAllString(expr, `${1}\.`)

		// If there is a final wildcard left replace it with an expression to match
		// any characters after the last dot.
		lc := len(expr) - 1
		if string(expr[lc]) == "*" {
			expr = expr[:lc] + ".*"
		}

		v.pattern, err = regexp.Compile(fmt.Sprintf("^%v$", expr))
		if err != nil {
			return err
		}
	}
	v.value = value
	return nil
}

// Match checks if current version matches a version string.
func (v Version) Match(version string) bool {
	if v.pattern != nil {
		return v.pattern.MatchString(version)
	}
	return v.value == version
}

// Resolve resolves a version pattern to a version string.
func (v Version) Resolve(versions []string) (string, error) {
	// Filter out versions that don't match current version pattern
	var valid []string
	for _, ver := range versions {
		// Append versions that match current version
		if v.Match(ver) {
			valid = append(valid, ver)
		}
	}

	if len(valid) == 0 {
		return "", &ErrVersionNotFound{Version: v.value}
	}
	sort.Slice(valid, compareVersions(valid))
	return valid[len(valid)-1], nil
}
