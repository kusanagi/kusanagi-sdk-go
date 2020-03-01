// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2020 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package transform

import (
	"reflect"

	"github.com/ugorji/go/codec"
)

// TODO: Implement a custom encoder/decoder for dates/times/datetimes/decimal

type Packable interface {
	Pack() ([]byte, error)
}

type Unpackable interface {
	Unpack(stream []byte) error
}

type Transformable interface {
	Packable
	Unpackable
}

// Pack serializes a value as binary.
func Pack(value interface{}) ([]byte, error) {
	var stream []byte

	// Prepare msgpack encoder
	handle := new(codec.MsgpackHandle)
	handle.WriteExt = true
	encoder := codec.NewEncoderBytes(&stream, handle)

	// Encode value to a stream
	err := encoder.Encode(value)
	if err != nil {
		return nil, err
	}
	return stream, nil
}

// Unpack deserializes a binary value to its original type.
func Unpack(stream []byte, value interface{}) error {
	// Prepare msgpack decoder
	handle := new(codec.MsgpackHandle)
	handle.MapType = reflect.TypeOf(map[string]interface{}(nil))
	handle.RawToString = true
	decoder := codec.NewDecoderBytes(stream, handle)

	// Decode the stream to a type
	if err := decoder.Decode(value); err != nil {
		return err
	}
	return nil
}
