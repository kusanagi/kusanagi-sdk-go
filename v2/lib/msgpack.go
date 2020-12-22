// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2020 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package lib

import (
	"bytes"
	"reflect"

	"github.com/ugorji/go/codec"
)

// Pack serializes a value as a msgpack binary.
func Pack(value interface{}) ([]byte, error) {
	handle := new(codec.MsgpackHandle)
	handle.WriteExt = true
	buffer := new(bytes.Buffer)
	encoder := codec.NewEncoder(buffer, handle)
	if err := encoder.Encode(value); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// Unpack a msgkpack binary value to its original type.
func Unpack(data []byte, value interface{}) error {
	handle := new(codec.MsgpackHandle)
	handle.MapType = reflect.TypeOf(map[string]interface{}(nil))
	handle.RawToString = true
	decoder := codec.NewDecoderBytes(data, handle)
	return decoder.Decode(value)
}
