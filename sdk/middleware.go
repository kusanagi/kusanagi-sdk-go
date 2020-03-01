// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2020 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package sdk

import (
	"errors"
	"fmt"
	"runtime"

	"github.com/kusanagi/kusanagi-sdk-go/payload"
)

// Constants used to recognize the type of middleware that should process a request.
const (
	requestMiddleware  = 1
	responseMiddleware = 2
)

// MiddlewareRequestCallback is called by middlewares when a service request is received.
type MiddlewareRequestCallback func(Request) (interface{}, error)

// MiddlewareResponseCallback is called by middlewares when a service response is received.
type MiddlewareResponseCallback func(Response) (Response, error)

// NewMiddleware creates a new KUSANAGI middleware component.
func NewMiddleware() *Middleware {
	// Get the source file name of the caller
	_, sourceFile, _, _ := runtime.Caller(1)
	m := Middleware{
		sourceFile: sourceFile,
		Component:  newComponent(),
	}
	m.Component.processCommand = m.processCommand
	return &m
}

// Middleware defines a KUSANAGI middleware component.
type Middleware struct {
	*Component

	request    MiddlewareRequestCallback
	response   MiddlewareResponseCallback
	sourceFile string
}

func (m *Middleware) processRequestCommand(args *payload.Payload, attrs map[string]string) (*payload.Payload, error) {
	// Get the meta data from the arguments
	meta := requestMeta{
		rid:             args.GetString("meta/id"),
		timestamp:       args.GetString("meta/datetime"),
		gatewayProtocol: args.GetString("meta/protocol"),
		clientAddr:      args.GetString("meta/client"),
		attributes:      attrs,
	}
	// Get the internal and public gateway addresses
	if items := args.GetSliceString("meta/gateway"); len(items) >= 2 {
		meta.gatewayInternalAddr = items[0]
		meta.gatewayAddr = items[1]
	} else {
		return nil, errors.New("failed to get gateway addresses from the payload meta")
	}

	// Get the call info from the arguments
	call := RequestCall{
		Service: args.GetString("call/service"),
		Version: args.GetString("call/version"),
		Action:  args.GetString("call/action"),
	}
	// Parse call parameters
	for _, data := range args.GetSliceMap("call/params") {
		pp := payload.NewParamFromMap(data)
		call.Params = append(call.Params, payloadToParam(pp))
	}
	// Create the HTTP request
	hrq, err := payloadToHttpRequest(payload.NewHttpRequestFromMap(args.GetMap("request")))
	if err != nil {
		return nil, err
	}

	// Create the request component
	req := newRequest(createAPI(m, m.sourceFile), &meta, &call, hrq)
	// Process the request and get the component
	c, err := m.request(req)
	if err != nil {
		m.triggerError(err)
		return nil, err
	}

	// For requests the component can be either a Request or a Response
	var result *payload.Payload
	if r, ok := c.(Request); ok {
		result = requestToResultPayload(r)
	} else if r, ok := c.(Response); ok {
		result = responseToResultPayload(r)
	} else {
		return nil, fmt.Errorf("unexpected return type for middleware callback: %T", c)
	}
	return result, nil
}

func (m *Middleware) processResponseCommand(args *payload.Payload, attrs map[string]string) (*payload.Payload, error) {
	meta := responseMeta{
		rid:             args.GetString("meta/id"),
		gatewayProtocol: args.GetString("meta/protocol"),
		gatewayAddr:     args.GetString("meta/gateway"),
		attributes:      attrs,
	}

	rv := ReturnValue{}
	if v, err := args.Get("return"); err == nil {
		rv.Set(v)
	}

	// Create the HTTP request
	hrq, err := payloadToHttpRequest(payload.NewHttpRequestFromMap(args.GetMap("request")))
	if err != nil {
		return nil, err
	}
	// Create the HTTP response
	hrs, err := payloadToHttpResponse(payload.NewHttpResponseFromMap(args.GetMap("response")))
	if err != nil {
		return nil, err
	}

	// Wrap the transport payload to allow read-only access to transport
	t := newTransport(payload.NewTransportFromMap(args.GetMap("transport")))

	// Create the response component
	var res Response
	res = newResponse(createAPI(m, m.sourceFile), &meta, rv, t, hrq, hrs)
	// Process the response and get the component
	res, err = m.response(res)
	if err != nil {
		m.triggerError(err)
		return nil, err
	}
	return responseToResultPayload(res), nil
}

func (m *Middleware) processCommand(name string, p *payload.Command) (*payload.CommandReply, error) {
	args := payload.New()
	args.Data = p.GetArgs()

	// Initialize request attributes with the existing values
	attrs := make(map[string]string)
	for name, v := range args.GetMap("meta/attributes") {
		attrs[name], _ = v.(string)
	}

	// Create the component and process the request or response
	var (
		err    error
		result *payload.Payload
	)
	switch args.GetInt64("meta/type") {
	case requestMiddleware:
		result, err = m.processRequestCommand(args, attrs)
		if err != nil {
			return nil, err
		}
	case responseMiddleware:
		result, err = m.processResponseCommand(args, attrs)
		if err != nil {
			return nil, err
		}
	default:
		v, _ := args.Get("meta/type")
		return nil, fmt.Errorf("unknown middleware type: %v", v)
	}

	// Add request attributes to the reply
	result.Set("attributes", attrs)

	return payload.NewCommandReply(name, result.Data), nil
}

// Request assigns a callback to execute when a service request is received.
func (m *Middleware) Request(c MiddlewareRequestCallback) {
	m.request = c
}

// Response assigns a callback to execute when a service response is received.
func (m *Middleware) Response(c MiddlewareResponseCallback) {
	m.response = c
}

func requestToResultPayload(r Request) *payload.Payload {
	call := payload.NewCall(r.GetServiceName(), r.GetServiceVersion(), r.GetActionName())

	// Parse parameters to payload params and add them to the result payload
	var pps []*payload.Param
	for _, p := range r.GetParams() {
		pps = append(pps, paramToPayload(p))
	}
	if len(pps) > 0 {
		call.SetParams(pps)
	}

	call.Entity()
	return call.Payload
}

func responseToResultPayload(r Response) *payload.Payload {
	hr := r.GetHttpResponse()
	hrp := payload.NewHttpResponse(hr.GetProtocolVersion(), hr.GetStatus(), string(hr.GetBody()))
	hrp.SetHeaders(hr.GetHeadersArray())
	hrp.Entity()
	return hrp.Payload
}
