// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2020 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package payload

// Types supported by the KUSANAGI framework
const TypeNull = "null"
const TypeBoolean = "boolean"
const TypeInteger = "integer"
const TypeFloat = "float"
const TypeString = "string"
const TypeBinary = "binary"
const TypeArray = "array"
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
