// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2020 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package sdk

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/kusanagi/kusanagi-sdk-go/payload"
	"github.com/kusanagi/kusanagi-sdk-go/protocol"
)

// Request represents a KUSANAGI framework request.
type Request interface {
	Api

	// GetID gets the request UUID.
	GetId() string

	// GetTimestamp gets request's creation timestamp.
	GetTimestamp() string

	// GetGatewayProtocol gets the protocol implemented by the gateway that is handling the request.
	GetGatewayProtocol() string

	// GetGatewayAddress gets the public address of the gateway handling the request.
	GetGatewayAddress() string

	// GetClientAddress gets the IP and port of the client that sent the request.
	GetClientAddress() string

	// SetAttribute sets a request attribute.
	// These attributes can only be read by `Response` objects.
	SetAttribute(name, value string) Request

	// GetServiceName gets the name of the service that should handle the request.
	GetServiceName() string

	// SetServiceName sets the name of the service that should handle the request.
	SetServiceName(name string) Request

	// GetServiceVersion gets the version of the service that should handle the request.
	GetServiceVersion() string

	// SetServiceVersion sets the version of the service that should handle the request.
	SetServiceVersion(version string) Request

	// GetActionName gets the name of the service action that should handle the request.
	GetActionName() string

	// SetActionName sets the name of the service action that should handle the request.
	SetActionName(action string) Request

	// HasParam checks if an parameter exists for current request.
	HasParam(name string) bool

	// GetParam gets a request parameter.
	// An empty parameter is returned when parameter doesn't exist.
	GetParam(name string) *Param

	// GetParams gets all the request parameters.
	GetParams() (ps []*Param)

	// SetParam adds a new parameter to the request.
	// If the parameter already exists it is overwritten.
	SetParam(p *Param) Request

	// NewParam creates a new parameter object.
	NewParam(name string, value interface{}, pType string) *Param

	// NewResponse creates a new response object.
	NewResponse(status int, text string) Response

	// GetHttpRequest gets an object containing the HTTP semantics of the request.
	GetHttpRequest() *HttpRequest
}

type requestMeta struct {
	rid                 string
	timestamp           string
	gatewayProtocol     string
	gatewayAddr         string
	gatewayInternalAddr string
	clientAddr          string
	attributes          map[string]string
}

type RequestCall struct {
	Service string
	Version string
	Action  string
	Params  []*Param
}

func newRequest(api *api, m *requestMeta, c *RequestCall, h *HttpRequest) *request {
	r := request{
		api:    api,
		meta:   m,
		http:   h,
		call:   c,
		params: make(map[string]*Param),
	}
	for _, p := range c.Params {
		r.params[p.GetName()] = p
	}
	return &r
}

type request struct {
	*api

	meta   *requestMeta
	http   *HttpRequest
	call   *RequestCall
	params map[string]*Param
}

func (r request) GetId() string {
	return r.meta.rid
}

func (r request) GetTimestamp() string {
	return r.meta.timestamp
}

func (r request) GetGatewayProtocol() string {
	return r.meta.gatewayProtocol
}

func (r request) GetGatewayAddress() string {
	return r.meta.gatewayAddr
}

func (r request) GetClientAddress() string {
	return r.meta.clientAddr
}

func (r *request) SetAttribute(name, value string) Request {
	r.meta.attributes[name] = value
	return r
}

func (r request) GetServiceName() string {
	return r.call.Service
}

func (r *request) SetServiceName(name string) Request {
	r.call.Service = name
	return r
}

func (r request) GetServiceVersion() string {
	return r.call.Version
}

func (r *request) SetServiceVersion(version string) Request {
	r.call.Version = version
	return r
}

func (r request) GetActionName() string {
	return r.call.Action
}

func (r *request) SetActionName(action string) Request {
	r.call.Action = action
	return r
}

func (r request) HasParam(name string) bool {
	_, ok := r.params[name]
	return ok
}

func (r request) GetParam(name string) *Param {
	if !r.HasParam(name) {
		return newEmptyParam(name)
	}
	return r.params[name]
}

func (r request) GetParams() (ps []*Param) {
	for _, p := range r.params {
		ps = append(ps, p)
	}
	return ps
}

func (r *request) SetParam(p *Param) Request {
	r.params[p.GetName()] = p
	return r
}

func (r request) NewParam(name string, value interface{}, pType string) *Param {
	return newParam(name, value, pType, true)
}

func (r request) NewResponse(status int, text string) Response {
	var hrs *HttpResponse

	// When protocol is HTTP create a default HTTP response
	hrq := r.GetHttpRequest()
	proto := r.GetGatewayProtocol()
	if proto == protocol.URN["http"] {
		if status == 0 {
			status = 200
		}
		if text == "" {
			text = "OK"
		}

		hrs = newHttpResponse(hrq.GetProtocolVersion(), status, text, nil, nil)
	}

	// Get the origin service when its available
	origin := []string{}
	if r.call.Service != "" && r.call.Version != "" && r.call.Action != "" {
		origin = append(origin, r.call.Service, r.call.Version, r.call.Action)
	}

	g := payload.GatewayAddr{
		Internal: r.meta.gatewayInternalAddr,
		Public:   r.GetGatewayAddress(),
	}
	tm := payload.NewTransportMeta(r.GetFrameworkVersion(), r.GetId(), r.GetTimestamp(), &g, origin, 0)
	t := newTransport(payload.NewTransport(tm))
	meta := responseMeta{
		rid:             r.GetId(),
		gatewayProtocol: proto,
		gatewayAddr:     r.GetGatewayAddress(),
		attributes:      r.meta.attributes,
	}
	return newResponse(r.api, &meta, ReturnValue{}, t, hrq, hrs)
}

func (r request) GetHttpRequest() *HttpRequest {
	return r.http
}

func newHttpRequest(
	method string,
	proto string,
	uri string,
	query url.Values,
	form url.Values,
	h http.Header,
	body []byte,
	fs []*File,
) (*HttpRequest, error) {
	u, err := url.ParseRequestURI(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTTP request URI: %v", err)
	}

	r := HttpRequest{
		method:       method,
		protoVersion: proto,
		url:          u,
		query:        query,
		form:         form,
		headers:      h,
		body:         body,
		files:        make(map[string]*File),
	}

	for _, f := range fs {
		r.files[f.GetName()] = f
	}
	return &r, nil
}

// HttpRequest provides read-only access to the HTTP semantics of a request.
type HttpRequest struct {
	method       string
	protoVersion string
	url          *url.URL
	query        url.Values
	form         url.Values
	headers      http.Header
	body         []byte
	files        map[string]*File
}

// IsMethod checks if a method matches the request's one.
// The HTTP method name is case insensitive.
func (hr HttpRequest) IsMethod(name string) bool {
	return hr.method == strings.ToUpper(name)
}

// GetMethod gets the name of the HTTP method in uppercase.
func (hr HttpRequest) GetMethod() string {
	return hr.method
}

// GetURL gets the URL for the request.
func (hr HttpRequest) GetUrl() string {
	return hr.url.String()
}

// GetURLScheme gets the URL scheme.
func (hr HttpRequest) GetUrlScheme() string {
	return hr.url.Scheme
}

// GetURLHost gets the URL host.
// When a port is given in the URL it will be added to host.
func (hr HttpRequest) GetUrlHost() string {
	return hr.url.Host
}

// GetURLPath gets the URL path.
func (hr HttpRequest) GetUrlPath() string {
	return hr.url.Path
}

// HasQueryParam checks if a param is defined in the query string.
func (hr HttpRequest) HasQueryParam(name string) bool {
	_, ok := hr.query[name]
	return ok
}

// GetQueryParam gets the value of an HTTP query parameter.
func (hr HttpRequest) GetQueryParam(name, defaultVal string) string {
	if !hr.HasQueryParam(name) {
		return defaultVal
	}
	return hr.query.Get(name)
}

// GetQueryParamArray gets the value of an HTTP query parameter as an array.
func (hr HttpRequest) GetQueryParamArray(name string, defaultVal []string) []string {
	if !hr.HasQueryParam(name) {
		return defaultVal
	}
	return hr.query[name]
}

// GetQueryParams gets all the HTTP query parameters.
func (hr HttpRequest) GetQueryParams() map[string]string {
	params := make(map[string]string)
	for name, values := range hr.query {
		params[name] = values[0]
	}
	return params
}

// GetQueryParamsArray gets all the HTTP query parameters.
// Each parameter value is returned as a slice.
func (hr HttpRequest) GetQueryParamsArray() map[string][]string {
	params := make(map[string][]string)
	for name, values := range hr.query {
		params[name] = values
	}
	return params
}

// HasPostParam checks if a param is defined in the request's form data.
func (hr HttpRequest) HasPostParam(name string) bool {
	_, ok := hr.form[name]
	return ok
}

// GetPostParam gets the value of an HTTP form data parameter.
func (hr HttpRequest) GetPostParam(name, defaultVal string) string {
	if !hr.HasPostParam(name) {
		return defaultVal
	}
	return hr.form.Get(name)
}

// GetPostParamArray gets the value of an HTTP form data parameter as an array.
func (hr HttpRequest) GetPostParamArray(name string, defaultVal []string) []string {
	if !hr.HasPostParam(name) {
		return defaultVal
	}
	return hr.form[name]
}

// GetPostParams gets all the HTTP form data parameters.
func (hr HttpRequest) GetPostParams() map[string]string {
	params := make(map[string]string)
	for name, values := range hr.form {
		params[name] = values[0]
	}
	return params
}

// GetPostParamsArray gets all the HTTP form data parameters.
// Each parameter value is returned as a slice.
func (hr HttpRequest) GetPostParamsArray() map[string][]string {
	params := make(map[string][]string)
	for name, values := range hr.form {
		params[name] = values
	}
	return params
}

// IsProtocolVersion checks if an HTTP protocol version matches the version of the request.
func (hr HttpRequest) IsProtocolVersion(version string) bool {
	return hr.protoVersion == version
}

// GetProtocolVersion gets the HTTP protocol version of the request.
func (hr HttpRequest) GetProtocolVersion() string {
	return hr.protoVersion
}

// HasHeader checks if an HTTP header exists.
func (hr HttpRequest) HasHeader(name string) bool {
	_, ok := hr.headers[name]
	return ok
}

// GetHeader gets the value of an HTTP header.
func (hr HttpRequest) GetHeader(name, defaultVal string) string {
	if !hr.HasHeader(name) {
		return defaultVal
	}
	return hr.headers.Get(name)
}

// GetHeaderArray gets the value of an HTTP header as an array.
func (hr HttpRequest) GetHeaderArray(name string, defaultVal []string) []string {
	if !hr.HasHeader(name) {
		return defaultVal
	}
	return hr.headers[name]
}

// GetHeaders gets all the HTTP headers.
func (hr HttpRequest) GetHeaders() map[string]string {
	headers := make(map[string]string)
	for name, values := range hr.headers {
		headers[name] = values[0]
	}
	return headers
}

// GetHeadersArray gets all the HTTP headers.
// Each parameter value is returned as a slice.
func (hr HttpRequest) GetHeadersArray() map[string][]string {
	headers := make(map[string][]string)
	for name, values := range hr.headers {
		headers[name] = values
	}
	return headers
}

// HasBody checks if the request body has content.
func (hr HttpRequest) HasBody() bool {
	return len(hr.body) > 0
}

// GetBody gets the contents of the HTTP request body.
func (hr HttpRequest) GetBody() []byte {
	return hr.body
}

// HasFile checks if a file was uploaded.
func (hr HttpRequest) HasFile(name string) bool {
	_, ok := hr.files[name]
	return ok
}

// GetFile gets an uploaded file.
// An empty file is returned when parameter doesn't exist.
func (hr HttpRequest) GetFile(name string) *File {
	if !hr.HasFile(name) {
		return newEmptyFile(name)
	}
	return hr.files[name]
}

// GetFiles gets all uploaded files.
func (hr HttpRequest) GetFiles() (fs []*File) {
	for _, f := range hr.files {
		fs = append(fs, f)
	}
	return fs
}

// Create a new HTTP request from an HTTP request payload.
func payloadToHttpRequest(p *payload.HttpRequest) (*HttpRequest, error) {
	var files []*File
	for _, fp := range p.GetFiles() {
		files = append(files, PayloadToFile(fp))
	}
	return newHttpRequest(
		p.GetMethod(),
		p.GetVersion(),
		p.GetURL(),
		p.GetQuery(),
		p.GetPostData(),
		p.GetHeaders(),
		p.GetBody(),
		files,
	)
}
