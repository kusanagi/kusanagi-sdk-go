// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2022 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package kusanagi

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kusanagi/kusanagi-sdk-go/v3/lib/payload"
)

func newResponse(c Component, s *state) *Response {
	return &Response{newApi(c, s)}
}

// Response API type for the middleware component.
type Response struct {
	*Api
}

// GetGatewayProtocol returns the protocol implemented by the gateway handling current request.
func (r *Response) GetGatewayProtocol() string {
	return r.command.Command.Arguments.Meta.Protocol
}

// GetGatewayAddress the public gateway address.
func (r *Response) GetGatewayAddress() string {
	return r.command.Command.Arguments.Meta.GetGateway()[1]
}

// GetRequestAttribute retuens a request attribute value.
//
// name: The attribute name.
// preset: A default value to use when the attribute doesn't exist.
func (r *Response) GetRequestAttribute(name, preset string) string {
	if v, exists := r.command.Command.Arguments.Meta.Attributes[name]; exists {
		return v
	}
	return preset
}

// GetRequestAttributes returns all the request attributes.
func (r *Response) GetRequestAttributes() map[string]string {
	return r.command.Command.Arguments.Meta.Attributes
}

// GetHTTPRequest returns the HTTP request semantics for the current response.
func (r *Response) GetHTTPRequest() *HTTPRequest {
	return newHTTPRequest(r.command.Command.Arguments.Request)
}

// GetHTTPResponse returns the HTTP response semantics for the current response.
func (r *Response) GetHTTPResponse() *HTTPResponse {
	rs := newHTTPResponse(r.reply.Command.Result.Response)
	rs.reply = r.reply
	return rs
}

// HasReturn checks if there is a return value.
//
// Return value is available when the initial service that is called
// has a return value, and returned a value in its command reply.
func (r *Response) HasReturn() bool {
	// TODO: See how to handle the null type for return values
	//       In the mean time the Go SDK won't support the null type.
	//       Maybe using some custom decoder for msgpack ?
	//       Return must be a struct with the value and an exists property.
	return r.command.Command.Arguments.Return != nil
}

// GetReturn returns the value returned by the called service.
func (r *Response) GetReturn() (interface{}, error) {
	if !r.HasReturn() {
		origin := r.command.Command.Arguments.Transport.Meta.Origin
		service := origin[0]
		version := origin[1]
		action := origin[2]
		err := fmt.Errorf(`No return value defined on "%s" (%s) for action: "%s"`, service, version, action)
		return nil, err
	}
	return r.command.Command.Arguments.Return, nil
}

// GetTransport returns the transport.
func (r *Response) GetTransport() Transport {
	return Transport{r.command.Command.Arguments.Transport}
}

func newHTTPResponse(p *payload.HTTPResponse) *HTTPResponse {
	r := HTTPResponse{
		payload: p,
		headers: make(map[string][]string),
	}

	// Index the headers using their upper case names
	for name, values := range p.Headers {
		r.headers[strings.ToUpper(name)] = values
	}

	return &r
}

// HTTPResponse represent an http response.
type HTTPResponse struct {
	payload *payload.HTTPResponse
	headers map[string][]string
	reply   *payload.Reply
}

// IsProtocolVersion checks if the response used the given HTTP version.
//
// version: The HTTP version.
func (r *HTTPResponse) IsProtocolVersion(version string) bool {
	return r.payload.Version == version
}

// GetProtocolVersion returns the HTTP version.
func (r *HTTPResponse) GetProtocolVersion() string {
	return r.payload.Version
}

// SetProtocolVersion sets the HTTP version.
//
// version: The HTTP version.
func (r *HTTPResponse) SetProtocolVersion(version string) *HTTPResponse {
	r.payload.Version = version
	if r.reply != nil {
		r.reply.Command.Result.Response.Version = r.payload.Version
	}
	return r
}

// IsStatus checks if the response uses the given status.
//
// status: The HTTP status.
func (r *HTTPResponse) IsStatus(status string) bool {
	return r.payload.Status == status
}

// GetStatus returns the HTTP status.
func (r *HTTPResponse) GetStatus() string {
	return r.payload.Status
}

// GetStatusCode returns the HTTP status code.
func (r *HTTPResponse) GetStatusCode() int {
	if code, err := strconv.Atoi(strings.SplitN(r.payload.Status, " ", 2)[0]); err == nil {
		return code
	}
	return 0
}

// GetStatusText returns the HTTP status text.
func (r *HTTPResponse) GetStatusText() string {
	if v := strings.SplitN(r.payload.Status, " ", 2); len(v) == 2 {
		return v[1]
	}
	return ""
}

// SetStatus sets the HTTP status code and text.
func (r *HTTPResponse) SetStatus(code int, text string) *HTTPResponse {
	r.payload.Status = fmt.Sprintf("%d %s", code, text)
	if r.reply != nil {
		r.reply.Command.Result.Response.Status = r.payload.Status
	}
	return r
}

// HasHeader checks if an HTTP header is defined.
//
// The header name is case insensitive.
//
// name: The HTTP header name.
func (r *HTTPResponse) HasHeader(name string) bool {
	_, exists := r.headers[strings.ToUpper(name)]
	return exists
}

// GetHeader returns an HTTP header.
//
// The header name is case insensitive.
//
// name: The HTTP header name.
// preset: A default value to use when the header doesn't exist.
func (r *HTTPResponse) GetHeader(name, preset string) string {
	if v, exists := r.headers[strings.ToUpper(name)]; exists && len(v) > 0 {
		return v[0]
	}
	return preset
}

// GetHeaderArray returns the HTTP header.
//
// The header name is case insensitive.
//
// name: The HTTP header name.
// preset: A default value to use when the header doesn't exist.
func (r *HTTPResponse) GetHeaderArray(name string, preset []string) []string {
	if v, exists := r.headers[strings.ToUpper(name)]; exists {
		return v
	}
	return preset
}

// GetHeaders returns all HTTP headers.
//
// The first value is returned for each header.
func (r *HTTPResponse) GetHeaders() map[string]string {
	headers := make(map[string]string)
	for name, values := range r.headers {
		headers[name] = values[0]
	}
	return headers
}

// GetHeadersArray returns all HTTP headers.
//
// Each header value is returned as a list.
func (r *HTTPResponse) GetHeadersArray() map[string][]string {
	headers := make(map[string][]string)
	for name, values := range r.headers {
		headers[name] = append([]string{}, values...)
	}
	return headers
}

// SetHeader sets an HTTP header with the given name and value.
//
// name: The HTTP header.
// value: The header value.
// overwrite: Allow existing headers to be overwritten.
func (r *HTTPResponse) SetHeader(name, value string, overwrite bool) *HTTPResponse {
	// If it exists get the original header name from the payload headers
	uppercaseName := strings.ToUpper(name)
	originalName := ""
	for headerName := range r.payload.Headers {
		if strings.ToUpper(headerName) == uppercaseName {
			originalName = headerName
			break
		}
	}

	// Initialize the headers map if it is empty
	if r.payload.Headers == nil {
		r.payload.Headers = make(map[string][]string)
	}

	// When a similar header exists replace the old header name with the new name
	// and add the new value. This can happen when the header name casing is different.
	if originalName != "" && originalName != name {
		// Create a new slice with the existing values if they should not be overwritten
		var values []string
		if !overwrite {
			values = append([]string{}, r.headers[uppercaseName]...)
		}
		// Append the new value to the list of headers
		values = append(values, value)
		// Remove the header with the previous name
		delete(r.payload.Headers, originalName)
		// Add the header with the new name
		r.payload.Headers[name] = values
	} else if overwrite {
		r.payload.Headers[name] = []string{value}
	} else {
		r.payload.Headers[name] = append(r.payload.Headers[name], value)
	}

	if r.reply != nil {
		r.reply.Command.Result.Response.Headers[name] = r.payload.Headers[name]
	}

	// Update the list of cached headers
	if _, exists := r.headers[uppercaseName]; exists && !overwrite {
		r.headers[uppercaseName] = append(r.headers[uppercaseName], value)
	} else {
		r.headers[uppercaseName] = []string{value}
	}

	return r
}

// HasBody checks if the HTTP response body has content.
func (r *HTTPResponse) HasBody() bool {
	return len(r.payload.Body) > 0
}

// GetBody returns the HTTP response body.
func (r *HTTPResponse) GetBody() []byte {
	return r.payload.Body
}

// SetBody sets the HTTP response body contents.
func (r *HTTPResponse) SetBody(content []byte) *HTTPResponse {
	r.payload.Body = content
	if r.reply != nil {
		r.reply.Command.Result.Response.Body = r.payload.Body
	}
	return r
}
