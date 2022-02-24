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
func Encode(value interface{}) ([]byte, error) {
	handle := new(codec.MsgpackHandle)
	handle.WriteExt = true
	buffer := new(bytes.Buffer)
	encoder := codec.NewEncoder(buffer, handle)
	if err := encoder.Encode(value); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// Decode a msgkpack binary value to its original type.
func Decode(data []byte, value interface{}) error {
	handle := new(codec.MsgpackHandle)
	handle.MapType = reflect.TypeOf(map[string]interface{}(nil))
	handle.RawToString = true
	decoder := codec.NewDecoderBytes(data, handle)
	return decoder.Decode(value)
}
