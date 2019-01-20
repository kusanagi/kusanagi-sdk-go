package transform

import "testing"

func TestPackAndUnpack(t *testing.T) {
	expected := map[string]interface{}{
		"field1":  42,
		"field22": []int{1, 33},
		"field2":  []string{"foo", "bar"},
		"field3":  map[string]int{"a": 1, "b": 2},
	}

	t.Log("pack a map with data")
	stream, err := Pack(expected)
	if err != nil {
		t.Fatal(err)
	}
	if stream == nil {
		t.Fatal("got an empty stream")
	}

	var value interface{}

	t.Log("unpack the stream back to a map")
	err = Unpack(stream, &value)
	if err != nil {
		t.Fatal(err)
	}

	original, ok := value.(map[string]interface{})
	if !ok {
		t.Fatal("invalid type unpacked")
	}

	if !AreEqual(original, expected) {
		t.Fatal("unpacked value doesn't match original value")
	}
}
