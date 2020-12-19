// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2020 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package kusanagi

// RequestCallback is called by middlewares when a service request is received.
//
// The result can be either a pointer to a Request or a Response component.
type RequestCallback func(*Request) (interface{}, error)

// ResponseCallback is called by middlewares when a service response is received.
type ResponseCallback func(*Response) (*Response, error)

// NewMiddleware creates a new Middleware component.
func NewMiddleware() *Middleware {
	return &Middleware{newComponent(middlewareRequestProcessor)}
}

// Middleware component.
type Middleware struct {
	component
}

// Request assigns a callback to execute when a service request is received.
//
// callback: Callback to handle requests.
func (m *Middleware) Request(callback RequestCallback) *Middleware {
	m.callbacks["request"] = callback
	return m
}

// Response assigns a callback to execute when a service response is received.
//
// callback: Callback to handle responses.
func (m *Middleware) Response(callback ResponseCallback) *Middleware {
	m.callbacks["response"] = callback
	return m
}
