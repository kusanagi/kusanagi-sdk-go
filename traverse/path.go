// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2020 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package traverse

import (
	"errors"
	"fmt"
	"strings"
)

// Sep defines the default character to use as path separator.
const Sep string = "/"

var (
	// ErrTraverseFailed defines an error for path traverse failures.
	ErrTraverseFailed = errors.New("path traverse failed")

	// ErrNotFound defines an error for values that are not found.
	ErrNotFound = errors.New("value not found")
)

// NewPath creates a new traverse path.
func NewPath(parts ...string) *Path {
	return &Path{parts: parts, Sep: Sep}
}

// NewSpacedPath creates a new traverse path that uses space character as separator.
func NewSpacedPath(parts ...string) *Path {
	return &Path{parts: parts, Sep: " "}
}

// TODO: Rewrite this module to use Path objects instead of path string and separator
// Path defines traversable paths.
type Path struct {
	parts []string
	Sep   string
}

func (p Path) String() string {
	return strings.Join(p.parts, p.Sep)
}

// Get gets a value from a map for a given path.
func Get(src map[string]interface{}, path, sep string, al *Aliases) (interface{}, error) {
	var (
		current interface{}
		exists  bool
		name    string
		ok      bool
		value   interface{}
	)

	if sep == "" {
		sep = Sep
	}
	parts := strings.Split(path, sep)
	last := len(parts) - 1
	item := src

	// Iterate all path elements until the last one to get its value
	for i, part := range parts {
		// When an empty space is used as separator it can happen that more than
		// one space is entered in the path by mistake, so handle the case by
		// skipping any invalid part generated by extra spaces.
		if part == "" {
			continue
		}

		// Get current part name
		if al == nil {
			name = part
		} else {
			name = al.Get(part)
		}

		// Get value for current path element
		current, exists = item[name]
		if !exists {
			return nil, ErrNotFound
		}

		// Stop iteration when the value is found
		if i == last {
			value = interface{}(current)
			break
		}

		// Value must be a map to be able to keep traversing path
		item, ok = current.(map[string]interface{})
		if !ok {
			return nil, ErrTraverseFailed
		}
	}
	return value, nil
}

// Exists check that a path exists inside a map.
func Exists(src map[string]interface{}, path, sep string, al *Aliases) bool {
	_, err := Get(src, path, sep, al)
	return err == nil
}

// Set sets a value in a map for a given path.
func Set(dst map[string]interface{}, path string, value interface{}, sep string, al *Aliases) error {
	var (
		current interface{}
		exists  bool
		name    string
	)

	if sep == "" {
		sep = Sep
	}
	parts := strings.Split(path, sep)
	last := len(parts) - 1
	item := dst

	// Iterate all path elements
	for i, part := range parts {
		// Get current part name
		if al == nil {
			name = part
		} else {
			name = al.Get(part)
		}

		// Set value when current path element is the last and finish loop
		if i == last {
			item[name] = value
			break
		}

		// Get value for current path element
		current, exists = item[name]
		if !exists {
			// A new map is created when current path element doesn't exists
			item[name] = make(map[string]interface{})
			current = item[name]
		}

		// Try to get current element value as a map
		item, _ = current.(map[string]interface{})
		// when value is not a map stop traversing and fail
		if item == nil {
			return ErrTraverseFailed
		}
	}
	return nil
}

// Delete deletes a value from a map for a given path.
func Delete(src map[string]interface{}, path, sep string, al *Aliases) error {
	if sep == "" {
		sep = Sep
	}
	parts := strings.Split(path, sep)
	last := len(parts) - 1

	// When path has one element delete it from src map
	if last == 0 {
		// Get part name alias when present
		name := parts[0]
		if al != nil {
			name = al.Get(parts[0])
		}
		delete(src, name)
		return nil
	}

	// Get path to parent element
	lastPart := parts[last]
	parentPath := strings.Join(parts[:last], sep)
	// Get parent element
	value, err := Get(src, parentPath, sep, al)
	if err != nil {
		return err
	}

	// Parent must be a map
	if _, ok := value.(map[string]interface{}); !ok {
		return ErrTraverseFailed
	}
	// Convert value to a map to be able to delete the value
	parent := value.(map[string]interface{})
	// Get past rt name alias when present
	if al != nil {
		lastPart = al.Get(lastPart)
	}
	delete(parent, lastPart)
	return nil
}

// Merge merges a map into another map, optionally extending slices.
func Merge(src map[string]interface{}, dst map[string]interface{}, al *Aliases, extend bool) error {
	var (
		name string
		err  error
	)

	for key, value := range src {
		name = key

		// Use name alias when aliases are available.
		// When a name has an alias it is used as key in the destination map.
		if al != nil {
			// Use name alias when current key doesn't exist in destination.
			if _, exists := dst[name]; !exists {
				name = al.Get(key)
			}
		}

		if _, exists := dst[name]; !exists { // Name doesn't exist in destination
			// Just add value
			dst[name] = value

		} else if s, ok := value.(map[string]interface{}); ok { // Merge maps
			// Get destination map for current name
			d, ok := dst[name].(map[string]interface{})
			if !ok {
				return fmt.Errorf("unsupported destination map type %T in merge", dst[name])
			}

			// Merge maps when name exist in destination and value is a map too
			if err = Merge(s, d, al, extend); err != nil {
				return err
			}
		} else if s, ok := value.([]interface{}); ok { // Merge slices
			// Get destination slice for current name
			d, ok := dst[name].([]interface{})
			if !ok {
				return fmt.Errorf("unsupported destination slice type %T in merge", dst[name])
			}

			// Merge slices
			dst[name] = append(d, s...)
		}
	}
	return nil
}
