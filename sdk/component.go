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
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	zmq "github.com/pebbe/zmq4"

	"github.com/kusanagi/kusanagi-sdk-go/cli"
	"github.com/kusanagi/kusanagi-sdk-go/logging"
	"github.com/kusanagi/kusanagi-sdk-go/payload"
	"github.com/kusanagi/kusanagi-sdk-go/protocol"
	"github.com/kusanagi/kusanagi-sdk-go/schema"
	"github.com/kusanagi/kusanagi-sdk-go/transform"
)

// ResourceFactory functions create resources to be stored in a component.
type ResourceFactory func(*Component) (interface{}, error)

// ErrorCallback function is called whenever an error is thrown or returned
// from a callback when processing a message in userland.
type ErrorCallback func(error)

// ComponentCallback is called for some events like "startup" or "shutdown".
type ComponentCallback func(*Component) error

// Defines a function to be called to process command payloads.
type processCommandFn func(string, *payload.Command) (*payload.CommandReply, error)

// Defines a function that parses a reply and returns the meta flags for the transport
type metaParserFn func(*payload.CommandReply) []byte

// Creates a new generic component.
func newComponent() *Component {
	// Parse command line arguments and init logging as soon as a component is created
	cli.Parse()
	logging.SetLevel(cli.Options.LogLevel)

	return &Component{
		resources: make(map[string]interface{}),
		registry:  schema.GetRegistry(),
	}
}

// Component defines a generic SDK component.
type Component struct {
	onStartup      ComponentCallback
	onShutdown     ComponentCallback
	onError        ErrorCallback
	processCommand processCommandFn
	resources      map[string]interface{}
	metaParser     metaParserFn
	registry       *schema.Registry
}

func (c Component) String() string {
	return fmt.Sprintf("\"%s\" (%s)", cli.Options.Name, cli.Options.Version)
}

func (c *Component) triggerStartup() error {
	if c.onStartup != nil {
		if err := c.onStartup(c); err != nil {
			return err
		}
	}
	return nil
}

func (c *Component) triggerShutdown() error {
	if c.onShutdown != nil {
		if err := c.onShutdown(c); err != nil {
			return err
		}
	}
	return nil
}

func (c *Component) triggerError(err error) {
	if c.onError != nil {
		c.onError(err)
	}
}

// HasResource checks if a resource name exist.
func (c Component) HasResource(name string) bool {
	_, ok := c.resources[name]
	return ok
}

// SetResource stores a resource.
func (c *Component) SetResource(name string, fn ResourceFactory) error {
	v, err := fn(c)
	if err != nil {
		return err
	} else if v == nil {
		return fmt.Errorf("invalid result value for resource: \"%s\"", name)
	}
	c.resources[name] = v
	return nil
}

// GetResource gets a resource by name.
func (c Component) GetResource(name string) (interface{}, error) {
	if v, ok := c.resources[name]; ok {
		return v, nil

	}
	return nil, fmt.Errorf("resource not found: \"%s\"", name)
}

// Startup registers a callback to be called during component startup.
func (c *Component) Startup(fn ComponentCallback) {
	c.onStartup = fn
}

// Shutdown registers a callback to be called during component shutdown.
func (c *Component) Shutdown(fn ComponentCallback) {
	c.onShutdown = fn
}

// Error registers a callback to be called on errors when processing a message in userland.
func (c *Component) Error(fn ErrorCallback) {
	c.onError = fn
}

// Log writes a string representation of a value value to the logs.
func (c Component) Log(value interface{}, level int) {
	if err := logging.DebugValue(level, value); err != nil {
		logging.Errorf("component value logging failed: %v", err)
	}
}

// Read the payload from stdin and process the action.
func (c *Component) processInput() error {
	if c.processCommand == nil {
		panic("command payload processor function is not assigned")
	}

	// Read the payload from the command line input
	logging.Debug("reading payload from stdin...")
	input, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		logging.Error(err)
		return errors.New("failed to read input")
	}

	// Deserialize the JSON input into a command payload
	p := payload.NewEmptyCommand()
	if err := transform.Deserialize(input, &p.Data); err != nil {
		logging.Error(err)
		return errors.New("failed to deserialize input payload")
	}

	// Call the component startup callback
	if err := c.triggerStartup(); err != nil {
		return err
	}

	// Process the command if it is a valid payload
	if !p.IsEmpty() {
		// Process the action payload and get a command reply
		reply, err := c.processCommand(cli.Options.Action, p)
		if err != nil {
			return err
		}
		// Serialize the reply as JSON
		output, err := transform.Serialize(reply.Data, true)
		if err != nil {
			logging.Error(err)
			return errors.New("failed to serialize command reply")
		}
		// Print response payload to stdout
		fmt.Println(string(output))
	}

	// Call the component shutdown callback
	if err := c.triggerShutdown(); err != nil {
		return err
	}
	return nil
}

func (c *Component) requestHandler(msg message) ([]byte, []byte, error) {
	// TODO: Add panic recover

	// Update the schema mappings
	if m := msg.Mappings(); len(m) != 0 {
		logging.Debug("updating schema mappings...")
		if err := c.registry.Update(m); err != nil {
			logging.Errorf("failed to update the schema mappings: %v", err)
		}
	}

	// Deserialize the message data into a command payload
	p := payload.NewEmptyCommand()
	if err := transform.Unpack(msg.Data(), &p.Data); err != nil {
		return nil, nil, fmt.Errorf("failed to deserialize payload: %v", err)
	}

	// Stop processing if command payload is not valid
	if p.IsEmpty() {
		return nil, nil, errors.New("request data doesn't contain a valid command payload")
	}

	// Process the action payload and get a command reply
	reply, err := c.processCommand(msg.Action(), p)
	if err != nil {
		return nil, nil, err
	}

	// Get meta flags for the response payload
	meta := []byte(string(protocol.EmptyMeta))
	if c.metaParser != nil {
		meta = c.metaParser(reply)
	}

	// Serialize the reply
	reply.Entity()
	output, err := reply.Pack()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to serialize command reply: %v", err)
	}
	return meta, output, nil
}

func (c *Component) workerFactory(requests chan [][]byte) *worker {
	return newWorker(requests, c.requestHandler)
}

func (c *Component) listen(address string) error {
	// Create a context to be able to terminate workers that use ZMQ sockets
	ctx, err := zmq.NewContext()
	if err != nil {
		return err
	}

	// Register a signal handler to terminate gracefully
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	go func() {
		// Block until a signal is received
		<-sigc
		// When the first signal is received stop listening for other signals
		signal.Stop(sigc)
		logging.Debug("termination signal received")
		// Terminate context to close all sockets and poller gracefully
		if err := ctx.Term(); err != nil {
			logging.Errorf("failed to terminate sockets: %v", err)
		} else {
			logging.Debug("sockets termnated successfully")
		}
	}()

	// Create a socket to receive incoming component requests
	incoming, err := ctx.NewSocket(zmq.ROUTER)
	if err != nil {
		return err
	}
	incoming.SetLinger(0)
	defer incoming.Close()

	// Start listening for incoming requests
	logging.Debugf("listening for requests at address: \"%s\"", address)
	if err := incoming.Bind(address); err != nil {
		if e := zmq.AsErrno(err); e == zmq.ETERM {
			return nil
		}
		return fmt.Errorf("failed to start listening for requests: %v", err)
	}
	defer incoming.Unbind(address)

	// Create an internal socket to receive worker responses
	responses, err := ctx.NewSocket(zmq.PULL)
	if err != nil {
		return err
	}
	responses.SetLinger(0)
	defer responses.Close()

	// Start listening for worker responses
	if err := responses.Bind("inproc://worker.responses"); err != nil {
		if e := zmq.AsErrno(err); e == zmq.ETERM {
			return nil
		}
		return err
	}

	// TODO: Tune the default values for workers and workload
	count, ok := cli.Options.Var.GetUint32("worker.count")
	if !ok {
		count = 10
	}
	load, ok := cli.Options.Var.GetUint32("worker.load")
	if !ok {
		load = 100
	}
	logging.Debugf("creating %d workers with a workload of %d each...", count, load)

	// Create a balancer to distribute requests among different workers
	balancer := newBalancer(count, load, c.workerFactory)
	if err := <-balancer.Start(ctx); err != nil {
		return err
	}
	// Before exit stop the balancer
	defer balancer.Stop()

	// Create a poller to handle sockets input and output
	poller := zmq.NewPoller()
	poller.Add(incoming, zmq.POLLIN)
	poller.Add(responses, zmq.POLLIN)
	logging.Info("component started")
MAIN:
	for {
		polled, err := poller.Poll(-1)
		if err != nil {
			// ETERM means the context has been terminated.
			// EINTR means a system interruption was triggered during a socket operation.
			errno := zmq.AsErrno(err)
			if errno == zmq.ETERM {
				// Finish main loop gracefully
				break MAIN
			} else if errno != zmq.Errno(syscall.EINTR) {
				logging.Errorf("socket poll failed: %v", err)
			}
			continue
		}

		for _, p := range polled {
			switch p.Socket {
			case incoming:
				// Read incoming data
				msg, err := incoming.RecvMessageBytes(0)
				if err != nil {
					if zmq.AsErrno(err) == zmq.ETERM {
						break MAIN
					} else {
						logging.Errorf("failed to read incoming request data: %v", err)
						continue
					}
				}
				// Queue the request to be processed by the worker with less workload
				// TODO: Implement execution timeout support (cli.Options.Timeout)
				balancer.Queue(msg)
			case responses:
				// Read response data from workers
				msg, err := responses.RecvMessageBytes(0)
				if err != nil {
					if zmq.AsErrno(err) == zmq.ETERM {
						break MAIN
					} else {
						logging.Errorf("failed to read workers data: %v", err)
						continue
					}
				}

				// Write response to main socket
				if _, err := incoming.SendMessage(msg); err != nil {
					if zmq.AsErrno(err) == zmq.ETERM {
						break MAIN
					} else {
						logging.Errorf("failed to write response data: %v", err)
						continue
					}
				}
			}
		}
	}
	logging.Info("component stopped")
	return nil
}

// Run starts the SDK component.
func (c *Component) Run() {
	logging.Debugf("running as PID: %d", os.Getpid())

	if cli.Options.DisableShortNames {
		payload.DisableShortNames = true
	}

	// Read payload from stdin when action name is available,
	// and after the payload is processed stop running.
	if cli.Options.Action != "" {
		if err := c.processInput(); err != nil {
			logging.Error(err)
			os.Exit(1)
		}
		// Stop running after the input payload is processed
		return
	}

	// Call the component startup callback
	logging.Info("running startup callback...")
	if err := c.triggerStartup(); err != nil {
		logging.Error(err)
		os.Exit(1)
	}

	// When no payload is available in stdin run the component
	// and start listening for requests.
	var address string
	if port := cli.Options.TCP; port > 0 {
		address = fmt.Sprintf("tcp://127.0.0.1:%d", port)
	} else {
		address = fmt.Sprintf("ipc://%s", cli.Options.Socket)
	}
	err := c.listen(address)
	if err != nil {
		logging.Error(err)
	}

	// Call the component shutdown callback.
	// This is triggered even when component listen fails.
	if err := c.triggerShutdown(); err != nil {
		logging.Error(err)
		os.Exit(1)
	}

	// Exit with error when component listen fails
	if err != nil {
		os.Exit(1)
	}
}

// Creates a new API object using the CLI option values.
func createAPI(c Resourcer, sourceFile string) *api {
	o := cli.Options
	a := newApi(c, sourceFile, o.Name, o.Version, o.FrameworkVersion)
	a.variables = o.Var.Values
	a.debug = o.Debug
	return a
}
