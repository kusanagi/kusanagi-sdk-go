// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2021 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package payload

// TypeNull defines the KUSANAGI type for null values.
const TypeNull = "null"

// TypeBoolean defines the KUSANAGI type for boolean values.
const TypeBoolean = "boolean"

// TypeInteger defines the KUSANAGI type for integer values.
const TypeInteger = "integer"

// TypeFloat defines the KUSANAGI type for float values.
const TypeFloat = "float"

// TypeString defines the KUSANAGI type for string values.
const TypeString = "string"

// TypeBinary defines the KUSANAGI type for binary values.
const TypeBinary = "binary"

// TypeArray defines the KUSANAGI type for array values.
const TypeArray = "array"

// TypeObject defines the KUSANAGI type for object values.
const TypeObject = "object"

var types = []string{
	TypeNull,
	TypeBoolean,
	TypeInteger,
	TypeFloat,
	TypeString,
	TypeBinary,
	TypeArray,
	TypeObject,
}

// IsValidType check is a type name is supported by the framework.
func IsValidType(name string) bool {
	for _, t := range types {
		if t == name {
			return true
		}
	}
	return false
}
