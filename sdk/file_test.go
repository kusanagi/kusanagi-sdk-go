package sdk

import (
	"path/filepath"
	"testing"
)

func TestFile(t *testing.T) {
	url := "http://localhost/file.txt"
	path := "testdata/file.txt"
	name := filepath.Base(path)
	mime := "plain/text"
	token := "ABC"
	size := uint64(8)

	// Empty file name is invalid
	if _, err := NewFile(" ", path, mime, name, size, token); err != ErrInvalidFileName {
		t.Error("expected an error for empty file name")
	}

	// HTTP file path without token is invalid
	if _, err := NewFile(name, url, mime, name, size, ""); err != ErrTokenRequired {
		t.Error("expected an error for HTTP file path without token")
	}

	// Non existing local files should fail
	if _, err := NewFile(name, "file://i-dont-exist.txt", mime, name, size, ""); err != nil {
		t.Error("expected an error for local files that don't exist")
	}

	// TODO: Continue testing the sdk file class
	// file, err := NewFile(name, path, mime, name, size, token)
	// if err != nil {
	// 	t.Fatal(err)
	// }
}
