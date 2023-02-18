// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2022 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package msgpack

import (
	"bytes"
	"reflect"

	"github.com/ugorji/go/codec"
)

// Encode serializes a value as a msgpack binary.
func Encode(v interface{}) ([]byte, error) {
	var (
		h   codec.MsgpackHandle
		buf bytes.Buffer
	)

	h.WriteExt = true

	enc := codec.NewEncoder(&buf, &h)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Decode a msgkpack binary value to its original type.
func Decode(b []byte, v interface{}) error {
	var h codec.MsgpackHandle

	h.MapType = reflect.TypeOf(map[string]interface{}(nil))
	h.RawToString = true

	dec := codec.NewDecoderBytes(b, &h)

	return dec.Decode(v)
}
