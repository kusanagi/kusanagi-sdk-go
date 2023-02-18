// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2023 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package payload

import "net/http"

// HTTPRequestData contains data sent in a request.
type HTTPRequestData map[string][]string

// HTTPRequest represents the semantics of an HTTP request.
type HTTPRequest struct {
	Version  string          `json:"v"`
	Method   string          `json:"m"`
	URL      string          `json:"u"`
	Query    HTTPRequestData `json:"q"`
	PostData HTTPRequestData `json:"p"`
	Headers  http.Header     `json:"h"`
	Body     []byte          `json:"b"`
	Files    []File          `json:"f"`
}
