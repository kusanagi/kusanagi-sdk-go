// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2022 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package kusanagi

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/kusanagi/kusanagi-sdk-go/v4/lib/payload"
)

func newRequest(c Component, s *state) *Request {
	api := newApi(c, s)

	// Index parameters by name
	params := make(map[string]payload.Param)
	for _, p := range api.reply.Command.Result.Call.Params {
		params[p.Name] = p
	}

	return &Request{api, params}
}

// Request API type for the middleware component.
type Request struct {
	*Api

	params map[string]payload.Param
}

// GetID returns the request UUID.
func (r *Request) GetID() string {
	return r.command.GetRequestID()
}

// GetTimestamp returns the request timestamp.
func (r *Request) GetTimestamp() string {
	return r.command.Command.Arguments.Meta.Datetime
}

// GetGatewayProtocol returns the protocol implemented by the gateway handling current request.
func (r *Request) GetGatewayProtocol() string {
	return r.command.Command.Arguments.Meta.Protocol
}

// GetGatewayAddress the public gateway address.
func (r *Request) GetGatewayAddress() string {
	return r.command.Command.Arguments.Meta.GetGateway()[1]
}

// GetClientAddress returns the IP address and port of the client which sent the request.
func (r *Request) GetClientAddress() string {
	return r.command.Command.Arguments.Meta.Client
}

// SetAttribute registers a request attribute.
func (r *Request) SetAttribute(name, value string) *Request {
	r.reply.Command.Result.Attributes[name] = value
	return r
}

// GetServiceName returns the name of the service.
func (r *Request) GetServiceName() string {
	return r.reply.Command.Result.Call.Service
}

// SetServiceName sets the name of the service.
//
// name: The name of the service.
func (r *Request) SetServiceName(name string) *Request {
	r.reply.Command.Result.Call.Service = name
	return r
}

// GetServiceVersion returns the version of the service.
func (r *Request) GetServiceVersion() string {
	return r.reply.Command.Result.Call.Version
}

// SetServiceVersion sets the version of the service.
//
// version: The version of the service.
func (r *Request) SetServiceVersion(version string) *Request {
	r.reply.Command.Result.Call.Version = version
	return r
}

// GetActionName returns the name of the action.
func (r *Request) GetActionName() string {
	return r.reply.Command.Result.Call.Action
}

// SetActionName sets the name of the action.
//
// name: The name of the action.
func (r *Request) SetActionName(name string) *Request {
	r.reply.Command.Result.Call.Action = name
	return r
}

// HasParam checks if a parameter exists.
//
// name: The name of the parameter.
func (r *Request) HasParam(name string) bool {
	_, exists := r.params[name]
	return exists
}

// GetParam returns a request parameter.
//
// name: The name of the parameter.
func (r *Request) GetParam(name string) *Param {
	if p, exists := r.params[name]; exists {
		return payloadToParam(p)
	}
	return newEmptyParam(name)
}

// GetParams returns all the request's parameters.
func (r *Request) GetParams() (params []*Param) {
	for _, p := range r.params {
		params = append(params, payloadToParam(p))
	}
	return params
}

// SetParam adds a new param for the current request.
//
// param: The parameter.
func (r *Request) SetParam(p *Param) *Request {
	payload := paramToPayload(p)
	r.params[p.GetName()] = payload
	r.reply.Command.Result.Call.Params = append(r.reply.Command.Result.Call.Params, payload)
	return r
}

// NewParam creates a new parameter.
//
// Creates an instance of Param with the given name, and optionally the value and data type.
// When the value is not provided then an empty string is assumed.
// If the data type is not defined then "string" is assumed.
//
// name: The parameter name.
// value: The parameter value.
// dataType: The data type of the value.
func (r *Request) NewParam(name string, value interface{}, dataType string) (*Param, error) {
	return newParam(name, value, dataType, true)
}

// NewResponse creates a new response.
//
// code: Optional status code.
// text: Optional status text.
func (r *Request) NewResponse(code int, text string) *Response {
	rs := newResponse(r.component, r.state)
	rs.GetHTTPResponse().SetStatus(code, text)
	// Change the reply payload from a request payload to a response payload.
	// Initially the reply for the Request component is a RequestPayload.
	r.reply.SetResponse(code, text)
	return rs
}

// GetHTTPRequest returns the HTTP request semantics for the current request.
func (r *Request) GetHTTPRequest() *HTTPRequest {
	return newHTTPRequest(r.command.Command.Arguments.Request)
}

func newHTTPRequest(p *payload.HTTPRequest) *HTTPRequest {
	r := HTTPRequest{
		payload: p,
		headers: make(map[string][]string),
		files:   make(map[string]File),
	}

	// Parse the URL and assign it to the request
	url, _ := url.Parse(p.URL)
	r.url = url

	// Index the headers using their upper case names
	for name, values := range p.Headers {
		r.headers[strings.ToUpper(name)] = values
	}

	// Index files by name
	for _, f := range p.Files {
		r.files[f.Name] = payloadToFile(&f)
	}

	return &r
}

// HTTPRequest represents an HTTP request.
type HTTPRequest struct {
	payload *payload.HTTPRequest
	headers map[string][]string
	url     *url.URL
	// TODO: Change this to make each file a list to support multiple files with same name
	files map[string]File
}

// IsMethod checks if the request used the given HTTP method.
//
// name: The HTTP method name.
func (r HTTPRequest) IsMethod(name string) bool {
	return r.payload.Method == strings.ToUpper(name)
}

// GetMethod returns the HTTP method.
func (r HTTPRequest) GetMethod() string {
	return strings.ToUpper(r.payload.Method)
}

// GetURL returns the request's URL
func (r HTTPRequest) GetURL() string {
	return r.url.String()
}

// GetURLScheme returns the URL scheme.
func (r HTTPRequest) GetURLScheme() string {
	return r.url.Scheme
}

// GetURLHost returns the URL hostname without port.
func (r HTTPRequest) GetURLHost() string {
	return r.url.Hostname()
}

// GetURLPort returns the URL port.
func (r HTTPRequest) GetURLPort() int {
	if v, err := strconv.Atoi(r.url.Port()); err == nil {
		return v
	}
	return 0
}

// GetURLPath returns the request's URL path.
func (r HTTPRequest) GetURLPath() string {
	return r.url.Path
}

// HasQueryParam checks if a param is defined in the HTTP query string.
//
// name: The HTTP param name.
func (r HTTPRequest) HasQueryParam(name string) bool {
	_, exists := r.payload.Query[name]
	return exists
}

// GetQueryParam returns the param value from the HTTP query string.
//
// The first value is returned when the parameter is present more
// than once in the HTTP query string.
//
// name: The HTTP param name.
// preset: A default value to use when the parmeter doesn't exist.
func (r HTTPRequest) GetQueryParam(name, preset string) string {
	// The query param value is always an array that contains the
	// actual parameter values. A param can have many values when
	// the HTTP string contains the parameter more than once.
	if v, exists := r.payload.Query[name]; exists && len(v) > 0 {
		return v[0]
	}
	return preset
}

// GetQueryParamArray returns the param value from the HTTP query string.
//
// The result is a list with all the values for the parameter.
// A parameter can be present more than once in an HTTP query string.
//
// name: The HTTP param name.
// preset: A default value to use when the parmeter doesn't exist.
func (r HTTPRequest) GetQueryParamArray(name string, preset []string) []string {
	if v, exists := r.payload.Query[name]; exists {
		return v
	}
	return preset
}

// GetQueryParams returns all HTTP query params.
//
// The first value of each parameter is returned when the parameter
// is present more than once in the HTTP query string.
func (r HTTPRequest) GetQueryParams() map[string]string {
	params := make(map[string]string)
	for name, values := range r.payload.Query {
		params[name] = values[0]
	}
	return params
}

// GetQueryParamsArray returns all HTTP query params.
//
// Each parameter value is returned as a list.
func (r HTTPRequest) GetQueryParamsArray() map[string][]string {
	params := make(map[string][]string)
	for name, values := range r.payload.Query {
		params[name] = append([]string{}, values...)
	}
	return params
}

// HasPostParam checks if a param is defined in the HTTP POST contents.
//
// name: The HTTP param name.
func (r HTTPRequest) HasPostParam(name string) bool {
	_, exists := r.payload.PostData[name]
	return exists
}

// GetPostParam returns the param value from the HTTP POST contents.
//
// The first value is returned when the parameter is present more
// than once in the HTTP request.
//
// name: The HTTP param name.
// preset: A default value to use when the parmeter doesn't exist.
func (r HTTPRequest) GetPostParam(name, preset string) string {
	if v, exists := r.payload.PostData[name]; exists && len(v) > 0 {
		return v[0]
	}
	return preset
}

// GetPostParamArray returns the param value from the HTTP POST contents.
//
// The result is a list with all the values for the parameter.
// A parameter can be present more than once in an HTTP request.
//
// name: The HTTP param name.
// preset: A default value to use when the parmeter doesn't exist.
func (r HTTPRequest) GetPostParamArray(name string, preset []string) []string {
	if v, exists := r.payload.PostData[name]; exists {
		return v
	}
	return preset
}

// GetPostParams returns all HTTP POST params.
//
// The first value of each parameter is returned when the parameter
// is present more than once in the HTTP request.
func (r HTTPRequest) GetPostParams() map[string]string {
	params := make(map[string]string)
	for name, values := range r.payload.PostData {
		params[name] = values[0]
	}
	return params
}

// GetPostParamsArray returns all HTTP POST params.
//
// Each parameter value is returned as a list.
func (r HTTPRequest) GetPostParamsArray() map[string][]string {
	params := make(map[string][]string)
	for name, values := range r.payload.PostData {
		params[name] = append([]string{}, values...)
	}
	return params
}

// IsProtocolVersion checks if the request used the given HTTP version.
//
// version: The HTTP version.
func (r HTTPRequest) IsProtocolVersion(version string) bool {
	return r.payload.Version == version
}

// GetProtocolVersion returns the HTTP version.
func (r HTTPRequest) GetProtocolVersion() string {
	return r.payload.Version
}

// HasHeader checks if an HTTP header is defined.
//
// The header name is case insensitive.
//
// name: The HTTP header name.
func (r HTTPRequest) HasHeader(name string) bool {
	_, exists := r.headers[strings.ToUpper(name)]
	return exists
}

// GetHeader returns an HTTP header.
//
// The header name is case insensitive.
//
// name: The HTTP header name.
// preset: A default value to use when the header doesn't exist.
func (r HTTPRequest) GetHeader(name, preset string) string {
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
func (r HTTPRequest) GetHeaderArray(name string, preset []string) []string {
	if v, exists := r.headers[strings.ToUpper(name)]; exists {
		return v
	}
	return preset
}

// GetHeaders returns all HTTP headers.
//
// The first value of each header is returned when the header
// is present more than once in the HTTP request.
func (r HTTPRequest) GetHeaders() map[string]string {
	headers := make(map[string]string)
	for name, values := range r.headers {
		headers[name] = values[0]
	}
	return headers
}

// GetHeadersArray returns all HTTP headers.
//
// Each header value is returned as a list.
func (r HTTPRequest) GetHeadersArray() map[string][]string {
	headers := make(map[string][]string)
	for name, values := range r.headers {
		headers[name] = append([]string{}, values...)
	}
	return headers
}

// HasBody checks if the HTTP request body has content.
func (r HTTPRequest) HasBody() bool {
	return len(r.payload.Body) > 0
}

// GetBody returns the HTTP request body.
func (r HTTPRequest) GetBody() []byte {
	return r.payload.Body
}

// HasFile checks if a file was uploaded in the current request.
//
// name: The name of the file parameter.
func (r HTTPRequest) HasFile(name string) bool {
	_, exists := r.files[name]
	return exists
}

// GetFile returns an uploaded file.
//
// name: The name of the file parameter.
func (r HTTPRequest) GetFile(name string) File {
	if f, exists := r.files[name]; exists {
		return f
	}
	return File{name: name}
}

// GetFiles returns all the uploaded files.
func (r HTTPRequest) GetFiles() (files []File) {
	for _, f := range r.files {
		files = append(files, f)
	}
	return files
}
