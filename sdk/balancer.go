// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2019 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package sdk

import (
	"container/heap"
	"errors"
	"fmt"

	"github.com/kusanagi/kusanagi-sdk-go/logging"
	"github.com/kusanagi/kusanagi-sdk-go/payload"
	zmq "github.com/pebbe/zmq4"
)

// ErrMaxWorkLoad is used to indicate that all component workers are at the maximum workload
// and they can't handle more requests.
var ErrMaxWorkLoad = errors.New("maximum work load reached")

// These defines the index for the message parts
const (
	mIdentity = iota
	mForwardIdentity
	mEmpty
	mRequestID
	mAction
	mMappings
	mPayload
)

// Message represents the multipart message received in KUSANAGI requests.
type message [][]byte

// ResponsePrefix returns the multipart prefix that has to be used for the responses.
func (msg message) ResponsePrefix() [][]byte {
	return [][]byte{msg.Identity(), msg.FwIdentity(), msg.Empty(), msg.RequestID()}
}

// Identity returns the ZMQ component socket identity.
func (msg message) Identity() []byte {
	return msg[mIdentity]
}

// FwIdentity returns the ZMQ KUSANAGI socket identity.
func (msg message) FwIdentity() []byte {
	return msg[mForwardIdentity]
}

// Empty returns an empty message part.
func (msg message) Empty() []byte {
	return msg[mEmpty]
}

// RequestID returns the ID for the current request.
func (msg message) RequestID() []byte {
	return msg[mRequestID]
}

// Action returns the name of the component action to process.
func (msg message) Action() string {
	return string(msg[mAction])
}

// Mappings returns the serialized mappings.
// Mappings are only present in the requests when they change, otherwise
// the returned value is nil.
func (msg message) Mappings() []byte {
	return msg[mMappings]
}

// Data returns the multipart data part with the serialized payload.
func (msg message) Data() []byte {
	return msg[mPayload]
}

// Done is used by the workers to signal when they finish processing a request.
type done struct {
	// Worker is the worker that finished processing a request
	Worker *worker

	// Error contains any error that was raised during request processing.
	// When an error is present it means request processing failed.
	Error error

	// Msg contains the response message ready to be written to a ZMQ socket.
	Msg [][]byte
}

// Request handler defines the function to call when a request arrives.
// The handler must process the request and return the response data
// already serialized, and also the meta flags describing the features
// that are enabled for the response payload.
type requestHandler func(m message) (meta []byte, data []byte, e error)

// Creates a new worker to handle KUSANAGI requests.
func newWorker(requests chan [][]byte, h requestHandler) *worker {
	return &worker{
		handler:  h,
		index:    -1, // -1 means is not inside a pool
		requests: requests,
	}
}

// A worker processes KUSANAGI requests to a component.
type worker struct {
	handler requestHandler
	// Count of jobs that are pending to process
	pending uint32
	// Heap index of the worker
	index int
	// Channel to receive the requests
	requests chan [][]byte
}

func (w *worker) SetHeapIndex(i int) {
	w.index = i
}

func (w *worker) GetHeapIndex() int {
	return w.index
}

func (w *worker) SetPending(v uint32) {
	w.pending = v
}

func (w *worker) GetPending() uint32 {
	return w.pending
}

// Queue queues a request message to be processed by the worker.
func (w *worker) Queue() chan [][]byte {
	return w.requests
}

func (w *worker) processRequest(msg message) ([][]byte, error) {
	// Call the handler to process the request message
	meta, data, err := w.handler(msg)
	if err != nil {
		// Create an error payload with the error message
		errp := payload.NewErrorFromObj(err)
		errp.Entity()
		// Serialize the error payload into a stream
		data, err = errp.Pack()
		if err != nil {
			return nil, fmt.Errorf("failed to serialize error payload: %v", err)
		}
	}
	// Return the multipart message with the response data
	return append(msg.ResponsePrefix(), meta, data), nil
}

func (w *worker) Start(d chan *done, quit chan struct{}) {
	if w.handler == nil {
		panic("a worker handler is not assinged")
	}

MAIN:
	for {
		select {
		case req := <-w.requests:
			msg, err := w.processRequest(req)
			d <- &done{Worker: w, Msg: msg, Error: err}
		case <-quit:
			break MAIN
		}
	}
}

// WorkerFactory defines a function to create new workers.
// During creation the worker receives a channel that has to be used to get
// all the queued requests.
type workerFactory func(requests chan [][]byte) *worker

// NewBalancer creates a new request balancer.
func newBalancer(workers, workload uint32, f workerFactory) *balancer {
	return &balancer{
		quit:     make(chan struct{}),
		pool:     pool{},
		done:     make(chan *done, workers),
		queue:    make(chan [][]byte, workers*100),
		worker:   f,
		workers:  workers,
		workload: workload,
	}
}

// Balancer is a request balancer that uses workers to process KUSANAGI requests.
type balancer struct {
	worker   workerFactory
	workers  uint32
	workload uint32
	quit     chan struct{}
	pool     pool
	done     chan *done
	queue    chan [][]byte
}

func (b *balancer) dispatch(msg [][]byte) error {
	// Get the worker that have less work load
	w := heap.Pop(&b.pool).(*worker)
	// Put worker back in the heap before exit
	defer heap.Push(&b.pool, w)
	// Assign the request to the worker
	select {
	case w.Queue() <- msg:
		w.SetPending(w.GetPending() + 1)
	default:
		// There is no more space in the worker's job queue because of high work load
		return ErrMaxWorkLoad
	}
	return nil
}

func (b *balancer) completed(w *worker) {
	w.SetPending(w.GetPending() - 1)
	heap.Remove(&b.pool, w.GetHeapIndex())
	// Put worker back in the heap
	heap.Push(&b.pool, w)
}

func (b *balancer) isRunning() bool {
	select {
	case <-b.quit:
		// When the channel is closed it means balancer is not running
		return false
	default:
	}
	return true
}

func (b *balancer) startWorkers() {
	count := int(b.workers)
	for i := 0; i < count; i++ {
		// Initialize a channel where the worker receives the request data
		requests := make(chan [][]byte, b.workload)
		// Create a worker using the worker factory
		w := b.worker(requests)
		// Finally start the worker in a different coroutine and add it to the pool
		go w.Start(b.done, b.quit)
		b.pool.Push(w)
	}
}

// Stop the balancer and its workers.
func (b *balancer) Stop() {
	if b.isRunning() {
		close(b.quit)
	}
}

// Queue adds a multipart request to the queue to be processed by a worker.
func (b *balancer) Queue(msg [][]byte) bool {
	select {
	case b.queue <- msg:
		return true
	default:
		// Ignore the request because the queue is full
	}
	return false
}

// Start creates the workers and starts the balancer.
func (b *balancer) Start(ctx *zmq.Context) chan error {
	b.startWorkers()
	errc := make(chan error)
	go func() {
		// Create a socket to forward the worker responses to the component server
		socket, err := ctx.NewSocket(zmq.PUSH)
		if err != nil {
			errc <- err
			return
		}
		socket.SetLinger(0)
		defer socket.Close()

		if err := socket.Connect("inproc://worker.responses"); err != nil {
			if errno := zmq.AsErrno(err); errno == zmq.ETERM {
				return
			}

			errc <- err
			return
		}
		// Close error channel to let the caller continue execution before
		// entering to the balancer main loop.
		close(errc)

	MAIN:
		for {
			select {
			case msg := <-b.queue:
				// When a request is in the queue forward it to the worker with less workload
				if err := b.dispatch(msg); err != nil {
					logging.Errorf("failed to process request: %v", err)
				}
			case done := <-b.done:
				// A worker finished its work
				b.completed(done.Worker)
				// If worker failed log the error, otherwise forward message to the router
				if done.Error != nil {
					logging.Errorf("worker error: %v", done.Error)
					continue
				} else if _, err := socket.SendMessage(done.Msg); err != nil {
					if errno := zmq.AsErrno(err); errno == zmq.ETERM {
						break MAIN
					} else {
						logging.Errorf("balancer failed to write response data: %v", err)
						continue
					}
				}
			case <-b.quit:
				// Stop balancer
				return
			}
		}
	}()
	return errc
}
