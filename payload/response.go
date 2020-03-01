// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2020 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package payload

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// NewEmptyHttpResponse creates a new empty HTTP response payload.
func NewEmptyHttpResponse() *HttpResponse {
	return &HttpResponse{Payload: NewNamespaced("response")}
}

// NewHttpResponse creates a new HTTP response payload.
func NewHttpResponse(version, status, body string) *HttpResponse {
	r := NewEmptyHttpResponse()
	r.SetBody(body)
	r.SetVersion(version)
	r.SetStatus(status)
	return r
}

func NewHttpResponseFromMap(m map[string]interface{}) *HttpResponse {
	r := NewEmptyHttpResponse()
	r.Data = m
	return r
}

// HttpResponse defines an HTTP response payload.
type HttpResponse struct {
	*Payload
}

// GetBody gets HTTP response body contents.
func (r HttpResponse) GetBody() string {
	return r.GetString("body")
}

// SetBody sets HTTP response body contents.
func (r *HttpResponse) SetBody(value string) error {
	return r.Set("body", value)
}

// GetVersion gets HTTP version of the response.
func (r HttpResponse) GetVersion() string {
	if s := r.GetString("version"); s != "" {
		return s
	}
	return "1.1"
}

// SetVersion sets HTTP version for the response.
func (r *HttpResponse) SetVersion(version string) error {
	return r.Set("version", version)
}

// GetStatus gets HTTP status code and text.
func (r HttpResponse) GetStatus() string {
	if s := r.GetString("status"); s != "" {
		return s
	}
	return "200 OK"
}

// GetStatusCode gets HTTP status code.
func (r HttpResponse) GetStatusCode() (int, error) {
	status := r.GetStatus()
	if parts := strings.SplitN(status, " ", 2); len(parts) >= 1 {
		if code, err := strconv.Atoi(parts[0]); err == nil {
			return code, nil
		}
	}
	return 0, fmt.Errorf("invalid HTTP response status: %s", status)
}

// GetStatusCode gets HTTP status text.
func (r HttpResponse) GetStatusText() string {
	if parts := strings.SplitN(r.GetStatus(), " ", 2); len(parts) == 2 {
		return parts[1]
	}
	return ""
}

// SetStatus sets HTTP status code and text.
func (r *HttpResponse) SetStatus(status string) error {
	return r.Set("status", status)
}

// GetHeaders gets HTTP response headers.
func (r HttpResponse) GetHeaders() http.Header {
	headers := http.Header{}
	if h := r.GetMap("headers"); h != nil {
		for name, v := range h {
			// Header values must be a slice of strings
			if values, ok := v.([]interface{}); ok {
				headers[name] = []string{}
				// Add header values skipping the ones that are not strings
				for _, v := range values {
					if value, ok := v.(string); ok {
						headers[name] = append(headers[name], value)
					}
				}
			}
		}
	}
	return headers
}

// SetHeaders sets HTTP response headers.
func (r *HttpResponse) SetHeaders(value map[string][]string) error {
	return r.Set("headers", value)
}
