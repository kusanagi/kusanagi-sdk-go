// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2021 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package payload

// Param represents an action param.
type Param struct {
	Name  string      `json:"n"`
	Value interface{} `json:"v,omitempty"`
	Type  string      `json:"t,omitempty"`
}
