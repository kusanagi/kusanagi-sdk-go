// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2023 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package semver

import (
	"fmt"
	"testing"
)

func TestZipVersionParts(t *testing.T) {
	tt := []struct {
		a        []string
		b        []string
		expected [][]string
	}{
		{[]string{"1", "2"}, []string{"3"}, [][]string{{"1", "3"}, {"2", ""}}},
		{[]string{"1"}, []string{"2", "3"}, [][]string{{"1", "2"}, {"", "3"}}},
		{[]string{"1"}, []string{"2", "3", "4"}, [][]string{{"1", "2"}, {"", "3"}, {"", "4"}}},
	}

	for _, tc := range tt {
		zip := zipVersionParts(tc.a, tc.b)
		if fmt.Sprint(zip) != fmt.Sprint(tc.expected) {
			t.Errorf("expected %v, got %v", tc.expected, zip)
		}
	}
}

func TestCompareEmpty(t *testing.T) {
	// Compare 2 empty values
	if compareEmpty("", "") {
		t.Error("comparing two empty values must be equal")
	}

	// Version with less parts is higher, which means
	// empty is higher than a non empty part.
	if !compareEmpty("A", "") {
		t.Error("empty must be higher than any other value")
	}
}

func TestCompareSubParts(t *testing.T) {
	tt := []struct {
		a       string
		b       string
		isLower bool
	}{
		// These values must be equal
		{"A", "A", false},
		{"1", "1", false},
		// First value is lower than the second one
		{"A", "B", true},
		{"1", "2", true},
		// First value is greater than second one
		{"B", "A", false},
		{"2", "1", false},
		// Integers are always lower than non integers
		{"A", "1", true},
		{"1", "A", false},
	}

	for i, tc := range tt {
		if lower := compareSubParts(tc.a, tc.b); lower != tc.isLower {
			t.Errorf("case %v, '%s' isLower:%v '%s' failed", i+1, tc.a, tc.isLower, tc.b)
		}
	}
}

func TestCompareVersions(t *testing.T) {
	tt := []struct {
		v1      string
		v2      string
		isLower bool
	}{
		{"A.B.C", "A.B", true},
		{"A.B-beta", "A.B", true},
		{"A.B-beta", "A.B-gamma", true},
		{"A.B.C", "A.B.C", false},
		{"A.B-alpha", "A.B-alpha", false},
		{"A.B", "A.B.C", false},
		{"A.B", "A.B-alpha", false},
		{"A.B-beta", "A.B-alpha", false},
	}

	for _, tc := range tt {
		versions := []string{tc.v1, tc.v2}
		if lower := compareVersions(versions); lower(0, 1) != tc.isLower {
			t.Errorf("'%s' isLower:%v '%s' failed", tc.v1, tc.isLower, tc.v2)
		}
	}
}

func TestMatch(t *testing.T) {
	// Check version match without wildcards
	version := New("1.2.3")
	if version.AllowWildcards() {
		t.Error(`expected version "1.2.3" not to support wildcards`)
	}

	if !version.Match("1.2.3") {
		t.Error("version match failed")
	}

	if version.Match("A.B.C") {
		t.Error("version matched with a different version")
	}

	// Check match with wildcards
	version = New("1.*.*")
	if !version.AllowWildcards() {
		t.Fatal(`expected version "1.*.*" to support wildcards`)
	}

	for _, v := range []string{"1.2.3", "1.4.3", "1.2.3-alpha"} {
		if !version.Match(v) {
			t.Fatalf(`wildcard version "%s" failed to match with "%s"`, version.value, v)
		}
	}

	for _, v := range []string{"A.B.C", "2.2.3"} {
		if version.Match(v) {
			t.Fatalf(`wildcard version "%s" matched with a different version: "%s"`, version.value, v)
		}
	}
}

func TestResolve(t *testing.T) {
	tt := []struct {
		pattern  string
		versions []string
		expected string
	}{
		{"*", []string{"3.4.0", "3.4.1", "3.4.a"}, "3.4.1"},
		{"3.*", []string{"3.4.0", "3.4.1", "3.4.a"}, "3.4.1"},
		{"3.4.1", []string{"3.4.0", "3.4.1", "3.4.a"}, "3.4.1"},
		{"3.4.*", []string{"3.4.0", "3.4.1", "3.4.a"}, "3.4.1"},
		{"3.4.*", []string{"3.4.a", "3.4.1", "3.4.0"}, "3.4.1"},
		{"3.4.*", []string{"3.4.alpha", "3.4.beta", "3.4.gamma"}, "3.4.gamma"},
		{"3.4.*", []string{"3.4.alpha", "3.4.a", "3.4.gamma"}, "3.4.gamma"},
		{"3.4.*", []string{"3.4.a", "3.4.12", "3.4.1"}, "3.4.12"},
		{"3.4.*", []string{"3.4.0", "3.4.0-a", "3.4.0-0"}, "3.4.0"},
		{"3.4.*", []string{"3.4.0-0", "3.4.0-a", "3.4.0-1"}, "3.4.0-1"},
		{"3.4.*", []string{"3.4.0-0", "3.4.0-1-0", "3.4.0-1"}, "3.4.0-1"},
	}

	for i, tc := range tt {
		version := New(tc.pattern)
		if v := version.Resolve(tc.versions); v != tc.expected {
			t.Errorf(`case %d expected "%s", got "%s"`, i+1, tc.expected, v)
		}
	}

	// A non matching pattern must give an error
	version := New("3.4.*.*")
	if v := version.Resolve([]string{"1.0", "A.B.C.D", "3.4.1"}); v != "" {
		t.Error("non matching patterns must return an empty string")
	}
}

func TestInvalidPattern(t *testing.T) {
	version := New("1.0.@")
	if version.AllowWildcards() {
		t.Error("invalid version pattern should not support wildcards")
	}
}
