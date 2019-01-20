package transform

import "testing"

func TestSerializeAndDeserialize(t *testing.T) {
	expected := map[string]interface{}{
		"field1": 42,
		"field2": []string{"foo", "bar"},
		"field3": map[string]int{"a": 1, "b": 2},
	}

	t.Log("serialize a map with data")
	stream, err := Serialize(expected, true)
	if err != nil {
		t.Fatal(err)
	}
	if stream == nil {
		t.Fatal("got an empty stream")
	}

	var value interface{}

	t.Log("deserialize the stream back to a map")
	err = Deserialize(stream, &value)
	if err != nil {
		t.Fatal(err)
	}

	original, ok := value.(map[string]interface{})
	if !ok {
		t.Fatal("invalid type deserialized")
	}

	if !AreEqual(original, expected) {
		t.Fatal("deserialized value doesn't match original value")
	}
}
