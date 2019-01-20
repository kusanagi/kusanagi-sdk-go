package traverse

import (
	"io/ioutil"
	"testing"

	"github.com/kusanagi/kusanagi-sdk-go/transform"
)

func TestGet(t *testing.T) {
	var expected int64 = 42

	src := map[string]interface{}{
		"f": map[string]interface{}{
			"b": expected,
		},
	}

	t.Log("find a value by path")
	value, err := Get(src, "f/b", Sep, nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("check that the value is correct")
	if value.(int64) != expected {
		t.Errorf("got the wrong value")
	}

	t.Log("check that traversing an invalid path fails")
	value, err = Get(src, "i/p", Sep, nil)
	if err == nil {
		t.Errorf("traversing didn't fail")
	}

	t.Log("check the traversing on a non map type fails")
	value, err = Get(src, "f/b/c", Sep, nil)
	if err != ErrTraverseFailed {
		t.Errorf("traversing didn't fail")
	}

	aliases := Aliases{
		"foo": "f",
		"bar": "b",
	}

	t.Log("find a value by path using aliases")
	value, err = Get(src, "foo/bar", Sep, &aliases)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("check that the value is correct")
	if value.(int64) != expected {
		t.Errorf("got the wrong value")
	}

	t.Log("check that traversing an invalid path fails using aliases")
	value, err = Get(src, "invalid/path", Sep, &aliases)
	if err == nil {
		t.Errorf("traversing didn't fail")
	}

	t.Log("check the traversing on a non map type fails using aliases")
	value, err = Get(src, "foo/bar/boom", Sep, &aliases)
	if err != ErrTraverseFailed {
		t.Errorf("traversing didn't fail")
	}
}

func TestExists(t *testing.T) {
	src := map[string]interface{}{
		"f": map[string]interface{}{
			"b": 42,
		},
	}

	t.Log("check that a path exists")
	if !Exists(src, "f/b", Sep, nil) {
		t.Error("path doesn't exists")
	}

	t.Log("check that a path doesn't exist")
	if Exists(src, "m/p", Sep, nil) {
		t.Error("path exists")
	}

	aliases := Aliases{
		"foo": "f",
		"bar": "b",
	}

	t.Log("check that a path exists using aliases")
	if !Exists(src, "foo/bar", Sep, &aliases) {
		t.Error("path doesn't exists")
	}

	t.Log("check that a path doesn't exist using aliases")
	if Exists(src, "missing/path", Sep, &aliases) {
		t.Error("path exists")
	}
}

func TestSet(t *testing.T) {
	var expected int64 = 42

	dst := make(map[string]interface{})
	path := "a/b"

	t.Log("check that a path doesn't exist")
	if Exists(dst, path, Sep, nil) {
		t.Error("path exists")
	}

	t.Log("set a value for a path that doesn't exist")
	err := Set(dst, path, expected, Sep, nil)
	if err != nil {
		t.Fatalf("set value failed with: %v", err)
	}

	t.Log("get the setted value")
	value, err := Get(dst, path, Sep, nil)
	if err != nil {
		t.Fatalf("get value failed with: %v", err)
	}

	t.Log("check that value is the right one")
	if value.(int64) != expected {
		t.Errorf("got the wrong value")
	}

	path = "foo/bar"
	aliases := Aliases{
		"foo": "f",
		"bar": "b",
	}

	t.Log("check that a path doesn't exist using aliases")
	if Exists(dst, path, Sep, &aliases) {
		t.Error("path exists")
	}

	t.Log("set a value for a path that doesn't exist using aliases")
	err = Set(dst, path, expected, Sep, &aliases)
	if err != nil {
		t.Fatalf("set value failed with: %v", err)
	}

	t.Log("get the setted value")
	value, err = Get(dst, path, Sep, &aliases)
	if err != nil {
		t.Fatalf("get value failed with: %v", err)
	}

	t.Log("check that value is the right one")
	if value.(int64) != expected {
		t.Errorf("got the wrong value")
	}

	t.Log("check that value is saved using short names")
	if !Exists(dst, "f/b", Sep, nil) {
		t.Fatalf("value is not saved using short names")
	}
}

func TestDelete(t *testing.T) {
	src := map[string]interface{}{
		"a": map[string]interface{}{
			"b": 42,
		},
	}
	path := "a/b"

	t.Log("check that path exist")
	if !Exists(src, path, Sep, nil) {
		t.Fatal("path doesn't exists")
	}

	t.Log("delete value for the path")
	err := Delete(src, path, Sep, nil)
	if err != nil {
		t.Fatalf("delete value failed with: %v", err)
	}

	t.Log("check that path doesn't exist")
	if Exists(src, path, Sep, nil) {
		t.Fatal("path still exists")
	}

	t.Log("deleting a path that doesn't exist must give an error")
	if Delete(src, "a/b/c", Sep, nil) == nil {
		t.Fatal("delete of missing path should have failed")
	}
}

func TestMerge(t *testing.T) {
	var (
		src interface{}
		dst interface{}
	)

	srcData, err := ioutil.ReadFile("testdata/source.json")
	if err != nil {
		t.Fatalf("failed to read data file: %s", err)
	}
	err = transform.Deserialize(srcData, &src)
	if err != nil {
		t.Fatalf("failed to deserialize data: %s", err)
	}

	dstData, err := ioutil.ReadFile("testdata/destination.json")
	if err != nil {
		t.Fatalf("failed to read data file: %s", err)
	}
	err = transform.Deserialize(dstData, &dst)
	if err != nil {
		t.Fatalf("failed to deserialize data: %s", err)
	}

	err = Merge(src.(map[string]interface{}), dst.(map[string]interface{}), nil, true)
	if err != nil {
		t.Fatal(err)
	}
}
