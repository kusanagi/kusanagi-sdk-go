// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2023 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package payload

import "net/http"

// HTTPResponseVersion defines the default HTTP version to use for responses.
const HTTPResponseVersion = "1.1"

// HTTPResponseStatus defines the default HTTP status to use for responses.
const HTTPResponseStatus = "200 OK"

// HTTPResponseContentType defines the default HTTP content type to use for responses.
const HTTPResponseContentType = "text/plain"

// NewHTTPResponse creates a new HTTP response payload.
func NewHTTPResponse() *HTTPResponse {
	h := http.Header{}
	h.Set("Content-Type", HTTPResponseContentType)

	return &HTTPResponse{
		Version: HTTPResponseVersion,
		Status:  HTTPResponseStatus,
		Headers: h,
		Body:    []byte(""),
	}
}

// HTTPResponse represents the semantics of an HTTP response.
type HTTPResponse struct {
	Version string      `json:"v"`
	Status  string      `json:"s"`
	Headers http.Header `json:"h"`
	Body    []byte      `json:"b"`
}

// GetVersion returns the HTTP version of the response.
func (r HTTPResponse) GetVersion() string {
	if r.Version == "" {
		return HTTPResponseVersion
	}
	return r.Version
}

// GetStatus returns the HTTP status code and text of the response.
func (r HTTPResponse) GetStatus() string {
	if r.Status == "" {
		return HTTPResponseStatus
	}
	return r.Status
}

// GetHeaders returns the HTTP headers of the response.
func (r HTTPResponse) GetHeaders() http.Header {
	if len(r.Headers) == 0 {
		h := http.Header{}
		h.Set("Content-Type", HTTPResponseContentType)
		return h
	}
	return r.Headers
}
