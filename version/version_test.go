package version

import (
	"fmt"
	"testing"
)

func TestZipVersionParts(t *testing.T) {
	cases := []struct {
		a        []string
		b        []string
		expected [][]string
	}{
		{[]string{"1", "2"}, []string{"3"}, [][]string{{"1", "3"}, {"2", ""}}},
		{[]string{"1"}, []string{"2", "3"}, [][]string{{"1", "2"}, {"", "3"}}},
		{[]string{"1"}, []string{"2", "3", "4"}, [][]string{{"1", "2"}, {"", "3"}, {"", "4"}}},
	}

	for _, tt := range cases {
		zip := zipVersionParts(tt.a, tt.b)
		if fmt.Sprint(zip) != fmt.Sprint(tt.expected) {
			t.Errorf("expected %v, got %v", tt.expected, zip)
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
	cases := []struct {
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

	for i, tt := range cases {
		lower := compareSubParts(tt.a, tt.b)
		if lower != tt.isLower {
			t.Errorf("case %v, '%s' isLower:%v '%s' failed", i+1, tt.a, tt.isLower, tt.b)
		}
	}
}

func TestCompareVersions(t *testing.T) {
	cases := []struct {
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

	for _, tt := range cases {
		versions := []string{tt.v1, tt.v2}
		lower := compareVersions(versions)
		if lower(0, 1) != tt.isLower {
			t.Errorf("'%s' isLower:%v '%s' failed", tt.v1, tt.isLower, tt.v2)
		}
	}
}

func TestMatchVersion(t *testing.T) {
	// Check version match without wildcards
	vs, err := New("1.2.3")
	if err != nil {
		t.Fatal(err)
	}

	if !vs.Match("1.2.3") {
		t.Error("version match failed")
	}

	if vs.Match("A.B.C") {
		t.Error("version matched with a different version")
	}

	// Check match with wildcards
	vs, err = New("1.*.*")
	if err != nil {
		t.Fatal(err)
	}

	for _, v := range []string{"1.2.3", "1.4.3", "1.2.3-alpha"} {
		if !vs.Match(v) {
			t.Fatalf("wildcard version %v failed to match with %v", vs.value, v)
		}
	}

	for _, v := range []string{"A.B.C", "2.2.3"} {
		if vs.Match(v) {
			t.Fatalf("wildcard version %v matched with a different version: %v", vs.value, v)
		}
	}
}

func TestResolveVersions(t *testing.T) {
	cases := []struct {
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

	for i, tt := range cases {
		vs, err := New(tt.pattern)
		if err != nil {
			t.Fatal(err)
		}

		version, err := vs.Resolve(tt.versions)
		if err != nil {
			t.Error(err)
		}

		if version != tt.expected {
			t.Errorf("case %v expected %s, got %s", i+1, tt.expected, version)
		}
	}

	// A non matching pattern must give an error
	vs, err := New("3.4.*.*")
	if err != nil {
		t.Fatal(err)
	}
	_, err = vs.Resolve([]string{"1.0", "A.B.C.D", "3.4.1"})
	if err == nil {
		t.Error("non matching patterns must return an error")
	}
}

func TestInvalidVersionResolveVersions(t *testing.T) {
	_, err := New("1.0.@")
	if err == nil {
		t.Error("invalid version pattern did't fail when creating version")
	}
}
