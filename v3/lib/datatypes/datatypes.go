// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2021 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package datatypes

import (
	"reflect"

	"github.com/kusanagi/kusanagi-sdk-go/v3/lib/payload"
)

// Null defines the KUSANAGI data type for null values.
const Null = payload.TypeNull

// Boolean defines the KUSANAGI data type for boolean values.
const Boolean = payload.TypeBoolean

// Integer defines the KUSANAGI data type for integer values.
const Integer = payload.TypeInteger

// Float defines the KUSANAGI data type for float values.
const Float = payload.TypeFloat

// String defines the KUSANAGI data type for string values.
const String = payload.TypeString

// Binary defines the KUSANAGI data type for binary values.
const Binary = payload.TypeBinary

// Array defines the KUSANAGI data type for array values.
const Array = payload.TypeArray

// Object defines the KUSANAGI data type for object values.
const Object = payload.TypeObject

// MaxUint defines the maximum size for an unsigned integer>
const MaxUint = ^uint(0)

// MinUint defines the minimum size for an unsigned integer.
const MinUint = 0

// MaxInt defines the maximum size for an integer.
const MaxInt = int(MaxUint >> 1)

// MinInt defines the minimum size for an integer.
const MinInt = -MaxInt - 1

// ResolveType resolves the type for a Go native data type.
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
	valueType := String

	kind := reflect.ValueOf(value).Kind()
	switch {
	case kind == reflect.String:
		valueType = String
	case kind == reflect.Bool:
		valueType = Boolean
	case reflect.Int <= kind && kind <= reflect.Int64:
		valueType = Integer
	case kind == reflect.Float32 || kind == reflect.Float64:
		valueType = Float
	case kind == reflect.Slice:
		valueType = Array
	case kind == reflect.Map:
		valueType = Object
	}
	return valueType
}
