// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2022 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package kusanagi

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kusanagi/kusanagi-sdk-go/v4/lib/cli"
	"github.com/kusanagi/kusanagi-sdk-go/v4/lib/log"
	"github.com/kusanagi/kusanagi-sdk-go/v4/lib/msgpack"
	"github.com/kusanagi/kusanagi-sdk-go/v4/lib/payload"
	"github.com/kusanagi/kusanagi-sdk-go/v4/lib/protocol"
	"github.com/pebbe/zmq4"
)

// State contains the context data for a multipart request of the framework.
type state struct {
	id      string
	action  string
	schemas payload.Mapping
	command payload.Command
	reply   *payload.Reply
	payload []byte
	input   cli.Input
	ctx     context.Context
	logger  log.RequestLogger
	request requestMsg
}

// Output for a request
type requestOutput struct {
	state    *state
	err      error
	response responseMsg
}

// Request processor processes ZMQ request messages for a component.
type requestProcessor func(*state, chan<- requestOutput)

// Create a response that contains an error as payload.
func createErrorResponse(message string) (responseMsg, error) {
	p := payload.NewErrorReply()
	p.Error.Message = message

	data, err := msgpack.Encode(p)
	if err != nil {
		return nil, err
	}

	return responseMsg{emptyFrame, data}, nil

}

// Pipe responses from a channel to a ZMQ internal socket
func pipeOutput(zctx *zmq4.Context, c <-chan requestOutput) error {
	errorc := make(chan error)

	go func() {
		// Create a socket to receive requests
		socket, err := zctx.NewSocket(zmq4.PAIR)
		if err != nil {
			errorc <- fmt.Errorf("Failed to create internal socket: %v", err)

			return
		}

		defer socket.Close()

		// Connect to the internal request forwarder
		if err := socket.Connect("inproc://responses"); err != nil {
			if errno := zmq4.AsErrno(err); errno != zmq4.ETERM {
				errorc <- fmt.Errorf("Failed to connect internal socket: %v", err)
			}

			return
		}

		// Close the socket after initialization
		close(errorc)

		// Start forwarding responses
		for output := range c {
			logger := output.state.logger
			response := output.response

			if output.err != nil {
				// Create an error response
				response, err = createErrorResponse(output.err.Error())
				if err != nil {
					// When the error response creation fails log the issue
					// and stop processing the response.
					logger.Errorf("Request failed with error: %v", output.err)
					logger.Errorf("Failed to create error response: %v", err)

					continue
				}
			}

			// Create the response message for the original request and send it to the forwarder
			msg := output.state.request.makeResponseMessage(response...)
			if _, err := socket.SendMessage([][]byte(msg)); err != nil {
				if zmq4.AsErrno(err) == zmq4.ETERM {
					break
				} else {
					log.Errorf("Failed to send internal response: %v", err)

					continue
				}
			}
		}
	}()

	// Wait until pipe initialization finishes
	return <-errorc
}

// Creates a new component server.
func newServer(input cli.Input, c Component, p requestProcessor) *server {
	return &server{c, input, p}
}

// SDK component server.
type server struct {
	component Component
	input     cli.Input
	processor requestProcessor
}

// Get the ZMQ channel address to use for listening incoming requests.
func (s *server) getAddress() (address string) {
	if s.input.IsTCPEnabled() {
		address = fmt.Sprintf("tcp://127.0.0.1:%d", s.input.GetTCP())
	} else if name := s.input.GetSocket(); name != "" {
		address = fmt.Sprintf("ipc://%s", name)
	} else {
		// Create a default name for the socket when no name is available.
		// The 'ipc://' prefix is removed from the string to get the socket name.
		address = protocol.IPC(s.input.GetComponent(), s.input.GetName(), s.input.GetVersion())
	}

	return address
}

func (s *server) hasComponentCallback(name string) bool {
	c := s.component.(*component)

	return c.hasCallback(name)
}

func (s *server) startMessageListener(msgc <-chan requestMsg) <-chan requestOutput {
	// Create a buffered channel to receive the responses from the handlers
	resc := make(chan requestOutput, 1000)

	// Handle messages until the messages channel is closed
	go func() {
		// TODO: See how to avoid race conditions when mapping are updated here (and read by userland)
		var schemas payload.Mapping

		// Get the title to use for the component
		title := s.input.GetComponentTitle()

		// Process execution timeout
		timeout := time.Duration(s.input.GetTimeout()) * time.Millisecond

		// Define a parent context for each request
		ctx, cancel := context.WithCancel(context.Background())

		for {
			// Block until a request message is received
			msg, ok := <-msgc
			if !ok {
				cancel()

				// When the channel is closed finish the loop
				break
			}

			// Check that the multipart message is valid
			if err := msg.check(); err != nil {
				log.Critical(err)

				// Log the error and continue listening for incoming requests
				continue
			}

			// Try to read the new schemas when present
			if v := msg.getSchemas(); v != nil {
				if err := msgpack.Decode(v, &schemas); err != nil {
					log.Errorf("Failed to read schemas: %v", err)
				}
			}

			// Process the request message in a new goroutine
			// TODO: Move to a function
			go func() {
				// Create a child context with the process execution timeout as limit
				ctx, cancel := context.WithTimeout(ctx, timeout)

				defer cancel()

				rid := msg.getRequestID()
				action := msg.getAction()
				logger := log.NewRequestLogger(rid)

				// State for the request
				state := state{
					id:      rid,
					action:  action,
					schemas: schemas,
					input:   s.input,
					ctx:     ctx,
					logger:  logger,
					request: msg,
				}

				// Prepare defaults for the request output
				output := requestOutput{state: &state}

				// Check that the request action is defined
				if !s.hasComponentCallback(msg.getAction()) {
					output.err = fmt.Errorf(`Invalid action for component %s: "%s"`, title, action)
					resc <- output

					return
				}

				// Try to read the new schemas when present
				if v := msg.getPayload(); v != nil {
					if err := msgpack.Decode(v, &state.command); err != nil {
						log.Criticalf("Failed to read payload: %v", err)

						output.err = fmt.Errorf(`Invalid payload for component %s: "%s"`, title, action)
						resc <- output

						return
					}
				} else {
					log.Critical("Empty command payload received")

					output.err = fmt.Errorf(`Empty command payload for component %s: "%s"`, title, action)
					resc <- output

					return
				}

				// Create a channel to wait for the processor output
				outc := make(chan requestOutput)

				// Process the request and return the response
				go s.processor(&state, outc)

				// Block until the processor finishes or the execution timeout is triggered
				select {
				case output := <-outc:
					resc <- output
				case <-ctx.Done():
					logger.Warningf("Execution timed out after %s. PID: %d", timeout, os.Getpid())
				}
			}()
		}
	}()

	return resc
}

func (s *server) start() error {
	// Define a custom ZMQ context
	zctx, err := zmq4.NewContext()
	if err != nil {
		return err
	}

	// Listen for termination signals
	go func() {
		// Define a channel to receive system signals
		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
		// Block until a signal is received
		<-sigc
		log.Debug("Termination signal received")
		// Terminate the ZMQ context to close sockets gracefully
		if err := zctx.Term(); err != nil {
			log.Errorf("Failed to terminate sockets context: %v", err)
		}
		// Clear the default ZMQ settings for retrying operations after EINTR.
		zmq4.SetRetryAfterEINTR(false)
		zctx.SetRetryAfterEINTR(false)
	}()

	// Create a socket to receive responses from the workers
	responses, err := zctx.NewSocket(zmq4.PAIR)
	if err != nil {
		return fmt.Errorf("Failed to create socket: %v", err)
	}
	defer responses.Close()

	// Make sure sockets close after context is terminated
	if err := responses.SetLinger(0); err != nil {
		return fmt.Errorf("Failed to set socket's linger option: %v", err)
	}

	// Start listenin from worker responses
	if err := responses.Bind("inproc://responses"); err != nil {
		return fmt.Errorf("Faled to open internal socket: %v", err)
	}
	defer responses.Unbind("inproc://responses")

	// Create a socket to receive incoming requests
	socket, err := zctx.NewSocket(zmq4.ROUTER)
	if err != nil {
		return fmt.Errorf("Failed to create socket: %v", err)
	}
	defer socket.Close()

	// Make sure sockets close after context is terminated
	if err := socket.SetLinger(0); err != nil {
		return fmt.Errorf("Failed to set socket's linger option: %v", err)
	}
	// Change the socket HWM to allow caching any number of incoming request.
	// ZMQ default value is 1000.
	if err := socket.SetRcvhwm(0); err != nil {
		return fmt.Errorf("Failed to set socket's high water mark option: %v", err)
	}

	// Start listening for incoming requests
	address := s.getAddress()
	log.Debugf(`Listening for request at address: "%s"`, address)
	if err := socket.Bind(address); err != nil {
		return fmt.Errorf(`Faled to open socket at address "%s": %v`, address, err)
	}
	defer socket.Unbind(address)

	// Create a buffered channel to send request payloads to the message listener.
	// The channel is buffered to allow faster request processing by the reactor.
	msgc := make(chan requestMsg, 1000)
	// On exit close the channel to avoid worker creation
	defer close(msgc)

	// Define a channel to read the responses from the processors.
	// The output is piped to be able to use send channel responses to the ZMQ socket
	if err := pipeOutput(zctx, s.startMessageListener(msgc)); err != nil {
		return err
	}

	// Create a poller to read and write sockets
	poller := zmq4.NewPoller()
	poller.Add(socket, zmq4.POLLIN)
	poller.Add(responses, zmq4.POLLIN)

MAIN:
	for {
		polled, err := poller.Poll(-1)
		if err != nil {
			// ETERM means the context has been terminated.
			// EINTR means a system interruption was triggered during a socket operation.
			errno := zmq4.AsErrno(err)
			if errno == zmq4.ETERM {
				break MAIN
			} else if errno != zmq4.Errno(syscall.EINTR) {
				log.Errorf("Socket poll failed: %v", err)
			}
			continue
		}

		for _, p := range polled {
			switch p.Socket {
			case socket:
				// Read the client request
				msg, err := socket.RecvMessageBytes(0)
				if err != nil {
					// When the context is terminated return the error to stop the reactor
					if zmq4.AsErrno(err) == zmq4.ETERM {
						break MAIN
					} else {
						log.Errorf("Failed to read request: %v", err)
						continue
					}
				}
				// Send the request to be processed by the workers
				msgc <- msg
			case responses:
				// Read the response from the internal socket
				msg, err := responses.RecvMessageBytes(0)
				if err != nil {
					if zmq4.AsErrno(err) == zmq4.ETERM {
						break MAIN
					} else {
						log.Errorf("Failed to read internal response: %v", err)
						continue
					}
				}

				// Write response to the client
				if _, err := socket.SendMessage(msg); err != nil {
					if zmq4.AsErrno(err) == zmq4.ETERM {
						break MAIN
					} else {
						log.Errorf("Failed to send response to client: %v", err)
						continue
					}
				}
			}
		}
	}

	log.Info("Component stopped")
	return nil
}
