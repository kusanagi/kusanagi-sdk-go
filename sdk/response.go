// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2019 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package sdk

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/kusanagi/kusanagi-sdk-go/payload"
)

// Response represents a KUSANAGI framework response.
type Response interface {
	Api

	// GetGatewayProtocol gets the protocol implemented by the gateway that is handling the request.
	GetGatewayProtocol() string

	// GetGatewayAddress gets the public address of the gateway handling the request.
	GetGatewayAddress() string

	// GetRequestAttribute gets a request attribute.
	// These attributes can only be setted by `Request` objects.
	GetRequestAttribute(name, defaultValue string) string

	// GetRequestAttribute gets all the request attributes.
	GetRequestAttributes() map[string]string

	// GetHttpRequest gets an object containing the HTTP semantics of the request.
	GetHttpRequest() *HttpRequest

	// GetHttpResponse gets an object containing the HTTP semantics of the response.
	GetHttpResponse() *HttpResponse

	// HasReturn checks whether or not the initial service call returned a value.
	HasReturn() bool

	// GetReturnValue gets the value returned by the called service.
	GetReturnValue() (interface{}, error)

	// GetTransport gets the read-only transport object.
	GetTransport() *Transport
}

func newErrNoReturnValueDefined(service, version, action string) error {
	return fmt.Errorf(`return value defined on "%s" (%s) for action: "%s"`, service, version, action)
}

type responseMeta struct {
	rid                 string
	gatewayProtocol     string
	gatewayAddr         string
	gatewayInternalAddr string
	attributes          map[string]string
}

func newResponse(
	api *api,
	m *responseMeta,
	rv ReturnValue,
	t *Transport,
	hrq *HttpRequest,
	hrs *HttpResponse,
) *response {
	return &response{
		api:          api,
		meta:         m,
		returnValue:  rv,
		transport:    t,
		httpRequest:  hrq,
		httpResponse: hrs,
	}
}

type response struct {
	*api

	meta         *responseMeta
	returnValue  ReturnValue
	transport    *Transport
	httpRequest  *HttpRequest
	httpResponse *HttpResponse
}

func (r response) GetGatewayProtocol() string {
	return r.meta.gatewayProtocol
}

func (r response) GetGatewayAddress() string {
	return r.meta.gatewayAddr
}

func (r response) GetRequestAttribute(name, defaultValue string) string {
	if v, ok := r.meta.attributes[name]; ok {
		return v
	}
	return defaultValue
}

func (r response) GetRequestAttributes() map[string]string {
	// A copy of the attributes map is returned to avoid modification of the original
	attrs := make(map[string]string)
	for k, v := range r.meta.attributes {
		attrs[k] = v
	}
	return attrs
}

func (r response) GetHttpRequest() *HttpRequest {
	return r.httpRequest
}

func (r response) GetHttpResponse() *HttpResponse {
	return r.httpResponse
}

func (r response) HasReturn() bool {
	return !r.returnValue.IsEmpty()
}

func (r response) GetReturnValue() (interface{}, error) {
	if !r.HasReturn() {
		o := r.transport.GetOriginService()
		return nil, newErrNoReturnValueDefined(o[0], o[1], o[2])
	}
	return r.returnValue.Get(), nil
}

func (r response) GetTransport() *Transport {
	return r.transport
}

func newHttpResponse(proto string, statusCode int, statusText string, h http.Header, body []byte) *HttpResponse {
	return &HttpResponse{
		protoVersion: proto,
		statusCode:   statusCode,
		statusText:   statusText,
		headers:      h,
		body:         body,
	}
}

// HttpResponse provides read-only access to the HTTP semantics of a response.
type HttpResponse struct {
	protoVersion string
	statusCode   int
	statusText   string
	headers      http.Header
	body         []byte
}

// IsProtocolVersion checks if an HTTP protocol version matches the version of the request.
func (hr HttpResponse) IsProtocolVersion(version string) bool {
	return hr.protoVersion == version
}

// GetProtocolVersion gets the HTTP protocol version of the request.
func (hr HttpResponse) GetProtocolVersion() string {
	return hr.protoVersion
}

// SetProtocolVersion sets the HTTP protocol version of the response.
func (hr *HttpResponse) SetProtocolVersion(version string) *HttpResponse {
	hr.protoVersion = version
	return hr
}

// IsStatus checks if a response status is the same as the one in the HTTP response.
// The HTTP status text is case insensitive.
func (hr HttpResponse) IsStatus(status string) bool {
	return strings.ToUpper(hr.GetStatus()) == strings.ToUpper(status)
}

// GetStatus gets the HTTP status of the response.
func (hr HttpResponse) GetStatus() string {
	return fmt.Sprintf("%d %s", hr.statusCode, hr.statusText)
}

// GetStatusCode gets the HTTP status code of the response.
func (hr HttpResponse) GetStatusCode() int {
	return hr.statusCode
}

// GetStatusText gets the HTTP status text of the response.
func (hr HttpResponse) GetStatusText() string {
	return hr.statusText
}

// SetStatus sets the HTTP status of the response.
func (hr *HttpResponse) SetStatus(code int, text string) *HttpResponse {
	hr.statusCode = code
	hr.statusText = text
	return hr
}

// HasHeader checks if an HTTP header exists.
func (hr HttpResponse) HasHeader(name string) bool {
	_, ok := hr.headers[name]
	return ok
}

// GetHeader gets the value of an HTTP header.
func (hr HttpResponse) GetHeader(name, defaultVal string) string {
	if !hr.HasHeader(name) {
		return defaultVal
	}
	return hr.headers.Get(name)
}

// GetHeaderArray gets the value of an HTTP header as an array.
func (hr HttpResponse) GetHeaderArray(name string, defaultVal []string) []string {
	if !hr.HasHeader(name) {
		return defaultVal
	}
	return hr.headers[name]
}

// GetHeaders gets all the HTTP headers.
func (hr HttpResponse) GetHeaders() map[string]string {
	headers := make(map[string]string)
	for name, values := range hr.headers {
		headers[name] = values[0]
	}
	return headers
}

// GetHeadersArray gets all the HTTP headers.
// Each parameter value is returned as a slice.
func (hr HttpResponse) GetHeadersArray() map[string][]string {
	headers := make(map[string][]string)
	for name, values := range hr.headers {
		headers[name] = values
	}
	return headers
}

// SetHeader sets the HTTP status of the response.
func (hr *HttpResponse) SetHeader(name, value string) *HttpResponse {
	hr.headers.Add(name, value)
	return hr
}

// HasBody checks if the request body has content.
func (hr HttpResponse) HasBody() bool {
	return len(hr.body) > 0
}

// GetBody gets the contents of the HTTP response body.
func (hr HttpResponse) GetBody() []byte {
	return hr.body
}

// SetBody sets the contents of the HTTP response body.
func (hr *HttpResponse) SetBody(body []byte) *HttpResponse {
	hr.body = body
	return hr
}

// Create a new HTTP response from an HTTP response payload.
func payloadToHttpResponse(p *payload.HttpResponse) (*HttpResponse, error) {
	code, err := p.GetStatusCode()
	if err != nil {
		return nil, err
	}

	hr := newHttpResponse(
		p.GetVersion(),
		code,
		p.GetStatusText(),
		p.GetHeaders(),
		[]byte(p.GetBody()),
	)
	return hr, nil
}
