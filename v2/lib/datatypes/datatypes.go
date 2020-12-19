// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2020 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package datatypes

import (
	"reflect"

	"github.com/kusanagi/kusanagi-sdk-go/v2/lib/payload"
)

// Types supported by the KUSANAGI framework
const Null = payload.TypeNull
const Boolean = payload.TypeBoolean
const Integer = payload.TypeInteger
const Float = payload.TypeFloat
const String = payload.TypeString
const Binary = payload.TypeBinary
const Array = payload.TypeArray
const Object = payload.TypeObject

// Maximum size for an unsigned integer
const MaxUint = ^uint(0)

// Minimum size for an unsigned integer
const MinUint = 0

// Maximum size for an integer
const MaxInt = int(MaxUint >> 1)

// Minimum size for an integer
const MinInt = -MaxInt - 1

// Resolves the type for a Go native data type.
//
// The resolved type is a KUSANAGI data type.
//
// value: The value from where to resolve the type name.
func ResolveType(value interface{}) string {
	if value == nil {
		return Null
	} else if _, ok := value.([]byte); ok {
		return Binary
	}

	// By default the type is string
	type_ := String

	kind := reflect.ValueOf(value).Kind()
	switch {
	case kind == reflect.String:
		type_ = String
	case kind == reflect.Bool:
		type_ = Boolean
	case reflect.Int <= kind && kind <= reflect.Int64:
		type_ = Integer
	case kind == reflect.Float32 || kind == reflect.Float64:
		type_ = Float
	case kind == reflect.Slice:
		type_ = Array
	case kind == reflect.Map:
		type_ = Object
	}
	return type_
}
