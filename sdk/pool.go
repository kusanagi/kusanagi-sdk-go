// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2019 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package sdk

// Poolable defines an interface for objects that can be added to a pool.
type poolable interface {
	SetHeapIndex(int)
	GetHeapIndex() int
	SetPending(uint32)
	GetPending() uint32
}

// Pool defines a "heap" to order poolable objects by pending jobs.
// It implements `heap.Interface`.
type pool []poolable

// Len returns the number of objects in the pool.
func (p pool) Len() int {
	return len(p)
}

// Less compares if an object has less pending jobs than other.
func (p pool) Less(i, j int) bool {
	return p[i].GetPending() < p[j].GetPending()
}

// Swap swaps the position of two objects by its position in the pool.
func (p pool) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
	p[i].SetHeapIndex(i)
	p[j].SetHeapIndex(j)
}

// Push adds an object to the pool.
func (p *pool) Push(v interface{}) {
	w := v.(poolable)
	w.SetHeapIndex(len(*p))
	*p = append(*p, w)
}

// Pop removes the last added object from the pool.
func (p *pool) Pop() interface{} {
	current := *p
	// Get the last worker from the heap
	size := len(current)
	w := current[size-1]
	// Change its index to -1 to "flag" that is no longer in a heap
	w.SetHeapIndex(-1)
	// Resize the heap to have all but the last element
	*p = current[0 : size-1]
	// And finally return the poped worker
	return w
}
