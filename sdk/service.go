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
	"github.com/kusanagi/kusanagi-sdk-go/protocol"
)

func transportMetaParser(rp *payload.CommandReply) []byte {
	// TODO: Implement. See: kusanagi/kusanagi-sdk-python3/kusanagi/service.py:52
	return []byte(string(protocol.EmptyMeta))
}

// ActionCallback is called when a service request is received.
type ActionCallback func(Action) (Action, error)

// NewService creates a new KUSANAGI service component.
func NewService() *Service {
	// Get the source file name of the caller
	_, sourceFile, _, _ := runtime.Caller(1)
	// Create the service component
	s := Service{
		actions:    make(map[string]ActionCallback),
		sourceFile: sourceFile,
		Component:  newComponent(),
	}
	s.Component.processCommand = s.processCommand
	s.Component.metaParser = transportMetaParser
	return &s
}

// Service defines a KUSANAGI service component.
type Service struct {
	*Component

	actions    map[string]ActionCallback
	sourceFile string
}

func (s *Service) processCommand(name string, p *payload.Command) (*payload.CommandReply, error) {
	// Get the callback to current action
	callback, ok := s.actions[name]
	if !ok {
		return nil, fmt.Errorf(`action not found: "%s"`, name)
	}

	// Create a new transport payload using the transport argument
	args := payload.New()
	args.Data = p.GetArgs()
	tp := payload.NewTransportFromMap(args.GetMap("transport"))
	if tp.IsEmpty() {
		return nil, errors.New("the transport is missing from the payload")
	}

	// Create a new action object
	rv := ReturnValue{}
	// TODO: Check that params is a slice (if so update SDK specs, otherwise check python SDK)
	action := newAction(createAPI(s, s.sourceFile), name, tp, &rv)
	action.setParams(args.GetSliceMap("params"))

	// Process the action
	if _, err := callback(action); err != nil {
		// Call the component error callback
		s.triggerError(err)
		return nil, err
	}
	result := payload.New()
	tp.Entity()
	result.Data = tp.Data
	if !rv.IsEmpty() {
		result.Set("return", rv.Get())
	}
	return payload.NewCommandReply(name, result.Data), nil
}

// Action assigns a callback to execute when a service action request is received.
func (s *Service) Action(name string, fn ActionCallback) {
	s.actions[name] = fn
}
