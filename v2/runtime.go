// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2021 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package kusanagi

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kusanagi/kusanagi-sdk-go/v2/lib/msgpack"
	"github.com/kusanagi/kusanagi-sdk-go/v2/lib/payload"
	"github.com/kusanagi/kusanagi-sdk-go/v2/lib/protocol"
	"github.com/kusanagi/kusanagi-sdk-go/v2/lib/runtime"
)

type callResult struct {
	ReturnValue interface{}
	Transport   *payload.Transport
	Duration    time.Duration
	Error       error
}

func call(
	ctx context.Context,
	address string,
	action string,
	callee []string,
	transport *payload.Transport,
	params []*Param,
	files []File,
	tcp bool,
	timeout uint,
) (<-chan callResult, error) {
	// Create the command payload arguments
	args := payload.CommandArguments{Transport: transport}
	args.SetAction(action)
	args.SetCallee(callee)

	if params != nil {
		args.Params = paramsToPayload(params)
	}

	if files != nil {
		args.Files = filesToPayload(files)
	}

	// Create the command payload for the call
	command := payload.NewCommand("runtime-call", "service")
	command.Command.Arguments = &args

	message, err := msgpack.Encode(command)
	if err != nil {
		return nil, fmt.Errorf("Failed to serialize the runtime call payload: %v", err)
	}

	// Make the runtime call
	c := make(chan callResult)
	go func() {
		// NOTE: Run-time calls are made to the server address where the caller is runnning
		//       and NOT directly to the service we wish to call. The KUSANAGI framework
		//       takes care of the call logic for us to keep consistency between all the SDKs.
		reply, duration, err := runtime.Call(ctx, protocol.SocketAddress(address, tcp), message, timeout)
		if err != nil {
			c <- callResult{Duration: duration, Error: err}
		} else if err := reply.Error; err != nil {
			c <- callResult{
				Duration: duration,
				Error:    errors.New(err.GetMessage()),
			}
		} else {
			c <- callResult{
				Duration:    duration,
				ReturnValue: reply.GetReturnValue(),
				Transport:   reply.GetTransport(),
			}
		}
		close(c)
	}()
	return c, nil
}
