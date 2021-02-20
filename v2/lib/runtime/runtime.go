// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2021 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package runtime

import (
	"context"
	"fmt"
	"time"

	"github.com/kusanagi/kusanagi-sdk-go/v2/lib/log"
	"github.com/kusanagi/kusanagi-sdk-go/v2/lib/msgpack"
	"github.com/kusanagi/kusanagi-sdk-go/v2/lib/payload"
	"github.com/pebbe/zmq4"
)

// Call makes a runtime call to a service.
func Call(ctx context.Context, address string, message []byte, timeout uint) (*payload.Reply, time.Duration, error) {
	var duration time.Duration

	// Define a custom ZMQ context
	zctx, err := zmq4.NewContext()
	if err != nil {
		return nil, duration, err
	}

	// Create a channel to stop waiting for parent's context done
	quit := make(chan struct{})
	defer close(quit)

	// When the context is done terminate the ZMQ context to stop the runtime call
	go func() {
		select {
		case <-ctx.Done():
			if err := zctx.Term(); err != nil {
				log.Errorf("Failed to terminate runtime call context: %v", err)
			}
		case <-quit:
		}
	}()

	// Create a socket to call the remote service
	socket, err := zctx.NewSocket(zmq4.REQ)
	if err != nil {
		return nil, duration, fmt.Errorf("Failed to create internal socket for runtime call: %v", err)
	}
	defer socket.Close()

	// Create a poller to be able to stop read on timeout
	poller := zmq4.NewPoller()
	poller.Add(socket, zmq4.POLLIN)

	// Connect to the local forwarder socket
	if err := socket.Connect(address); err != nil {
		return nil, duration, fmt.Errorf("Failed to connect to the forwarder socket: %v", err)
	}

	// Send the payload
	start := time.Now()
	if _, err := socket.SendMessage([]byte("\x01"), message); err != nil {
		return nil, duration, fmt.Errorf("Failed to send runtime call message: %v", err)
	}

	// Wait for the response
	if _, err := poller.PollAll(time.Duration(timeout) * time.Millisecond); err != nil {
		duration = time.Since(start) * time.Millisecond
		return nil, duration, fmt.Errorf("Failed to poll runtime call reply: %v", err)
	}

	// Read response
	response, err := socket.RecvBytes(0)
	if err != nil {
		duration = time.Since(start) * time.Millisecond
		return nil, duration, fmt.Errorf("Failed to read runtime call response: %v", err)
	}

	// Set call duration when the response is received
	duration = time.Since(start) * time.Millisecond

	var reply *payload.Reply
	if err := msgpack.Decode(response, &reply); err != nil {
		return nil, duration, fmt.Errorf("Failed to parse runtime call response: %v", err)
	}
	return reply, duration, nil
}
