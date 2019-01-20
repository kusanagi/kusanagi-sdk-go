// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2019 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package payload

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/kusanagi/kusanagi-sdk-go/transform"
	"github.com/kusanagi/kusanagi-sdk-go/traverse"
)

// TODO: CHange functions to use this type instead of map
type Data map[string]interface{}

type Mergeable interface {
	GetData() Data
}

// New creates a new payload.
func New() *Payload {
	return &Payload{
		Aliases: Aliases,
		Data:    Data{},
	}
}

// NewNamespaced creates a new payload with a data namespace.
func NewNamespaced(namespace string) *Payload {
	p := New()
	p.namespace = namespace
	return p
}

// NewFromMap creates a new payload from a map.
func NewFromMap(namespace string, data map[string]interface{}) *Payload {
	p := NewNamespaced(namespace)
	if data != nil {
		p.Data = data
	}
	return p
}

// Payload stores payload data.
type Payload struct {
	namespace string
	Data      map[string]interface{}
	Aliases   traverse.Aliases `json:"-"` // Don't serialize the aliases
}

// Size gets the payload size.
func (p Payload) Size() int {
	return len(p.Data)
}

// GetNamespace gets the payload data namespace.
func (p Payload) GetNamespace() string {
	return p.namespace
}

// GetData gets the payload data.
func (p Payload) GetData() map[string]interface{} {
	return p.Data
}

// IsEmpty checks if payload is empty.
func (p Payload) IsEmpty() bool {
	return p.Size() == 0
}

// GetAliases gets aliases to use for payload field names.
func (p Payload) GetAliases() *traverse.Aliases {
	// Don't use field aliases when short property names are disabled
	if DisableShortNames {
		return nil
	}

	// By default get the aliases defined in current module
	return &p.Aliases
}

// Entity transform payload data to an entity.
func (p *Payload) Entity() {
	// When there is no entity name leave data unmodified
	if p.namespace == "" {
		return
	}

	// Add entity namespace to payload data
	name := p.GetAliases().Get(p.namespace)
	p.Data = map[string]interface{}{
		name: p.Data,
	}
}

// UndoEntity removes entity namespace from paylaod data.
func (p *Payload) UndoEntity() bool {
	// When there is no entity name leave data unmodified
	if p.namespace == "" {
		return false
	}

	v, err := traverse.Get(p.Data, p.namespace, "", p.GetAliases())
	if err != nil {
		return false
	}
	data, ok := v.(map[string]interface{})
	if !ok {
		return false
	}
	p.Data = data
	return true
}

// Get a value from the payload using a custom separator.
func (p *Payload) Pget(path, sep string) (interface{}, error) {
	if p.Data == nil {
		return nil, errors.New("payload is empty")
	}

	return traverse.Get(p.Data, path, sep, p.GetAliases())
}

// PgetDefault gets a value from the payload using a custom separator or return a default when the path doesn't exist.
func (p *Payload) PgetDefault(path, sep string, defaultValue interface{}) interface{} {
	if p.Data == nil {
		return defaultValue
	}

	value, err := p.Pget(path, sep)
	if err != nil {
		return defaultValue
	}
	return value
}

// Get a value from the payload.
func (p *Payload) Get(path string) (interface{}, error) {
	return p.Pget(path, traverse.Sep)
}

// GetDefault gets a value from the payload or return a default when the path doesn't exist.
func (p *Payload) GetDefault(path string, defaultValue interface{}) interface{} {
	return p.PgetDefault(path, traverse.Sep, defaultValue)
}

func (p Payload) GetString(path string) string {
	if v, ok := p.GetDefault(path, "").(string); ok {
		return v
	}
	return ""
}

func (p Payload) GetBool(path string) bool {
	if v, ok := p.GetDefault(path, false).(bool); ok {
		return v
	}
	return false
}

func (p Payload) GetInt(path string) int {
	return p.GetDefault(path, 0).(int)
}

func (p Payload) GetInt64(path string) int64 {
	return p.GetDefault(path, int64(0)).(int64)
}

func (p Payload) GetUint64(path string) uint64 {
	value := p.GetDefault(path, uint64(0))
	if v, ok := value.(uint64); ok {
		return v
	}
	// TODO: Implement pre casting to string for the other type getters
	if v, err := strconv.ParseUint(fmt.Sprintf("%v", value), 10, 64); err == nil {
		return v
	}
	return 0
}

func (p Payload) GetFloat64(path string) float64 {
	return p.GetDefault(path, float64(0.0)).(float64)
}

func (p Payload) GetSlice(path string) []interface{} {
	if v := p.GetDefault(path, nil); v != nil {
		return v.([]interface{})
	}
	return nil
}

func (p Payload) GetSliceMap(path string) (values []map[string]interface{}) {
	for _, v := range p.GetSlice(path) {
		if value, ok := v.(map[string]interface{}); ok {
			values = append(values, value)
		}
	}
	return values
}

func (p Payload) GetSliceString(path string) (values []string) {
	for _, v := range p.GetSlice(path) {
		if value, ok := v.(string); ok {
			values = append(values, value)
		}
	}
	return values
}

func (p Payload) PgetSlice(path, sep string) []interface{} {
	if v, ok := p.PgetDefault(path, sep, nil).([]interface{}); ok {
		return v
	}
	return nil
}

func (p Payload) GetMap(path string) map[string]interface{} {
	if v, ok := p.GetDefault(path, nil).(map[string]interface{}); ok {
		return v
	}
	return nil
}

func (p Payload) PgetMap(path, sep string) map[string]interface{} {
	if v, ok := p.PgetDefault(path, sep, nil).(map[string]interface{}); ok {
		return v
	}
	return nil
}

// Exists checks if a path exists in the payload using a custom separator.
func (p Payload) Pexists(path, sep string) bool {
	return traverse.Exists(p.Data, path, sep, p.GetAliases())
}

// Exists checks if a path exists in the payload.
func (p Payload) Exists(path string) bool {
	return p.Pexists(path, traverse.Sep)
}

// Set a value in the payload using a custom separator.
func (p *Payload) Pset(path string, value interface{}, sep string) error {
	return traverse.Set(p.Data, path, value, sep, p.GetAliases())
}

// Set a value in the payload.
func (p *Payload) Set(path string, value interface{}) error {
	return p.Pset(path, value, traverse.Sep)
}

// Delete a value from the payload using a custom separator.
func (p *Payload) Pdelete(path, sep string) error {
	return traverse.Delete(p.Data, path, sep, p.GetAliases())
}

// Delete a value from the payload.
func (p *Payload) Delete(path string) error {
	return p.Pdelete(path, traverse.Sep)
}

// Merge merges data into the payload.
func (p *Payload) Merge(src Mergeable, aliases bool) error {
	var a *traverse.Aliases
	if aliases {
		a = p.GetAliases()
	}
	return traverse.Merge(src.GetData(), p.Data, a, true)
}

// MergeMap merges a map into the payload.
func (p *Payload) Pmerge(path string, data map[string]interface{}, sep string, aliases bool) error {
	var dst map[string]interface{}

	if p.Pexists(path, sep) {
		// The value in the path must be a map
		dst = p.PgetMap(path, sep)
		if dst == nil {
			return fmt.Errorf("value in path \"%s\" is not traversable", path)
		}
	} else {
		// When the path is not available in current payload create
		// an empty map to be able to merge the data.
		dst = make(map[string]interface{})
		p.Pset(path, dst, sep)
	}

	var a *traverse.Aliases
	if aliases {
		a = p.GetAliases()
	}
	return traverse.Merge(data, dst, a, true)
}

// Ppush pushes a value using a custom separator.
func (p *Payload) Ppush(path string, value interface{}, sep string) error {
	aliases := p.GetAliases()
	current := p.Data
	parts := strings.Split(path, sep)
	lastPartPos := len(parts) - 1
	for i, name := range parts {
		// Resolve alias for the path parts that doesn't exist in payload
		if _, exists := current[name]; !exists {
			name = aliases.Get(name)
		}

		// Current part is the last path part
		if i == lastPartPos {
			v, ok := current[name]
			if !ok {
				// When the last part doesn't exist create an empty list
				current[name] = []interface{}{}
				// Update current value pointer with the newly created list
				v = current[name]
			}

			items, ok := v.([]interface{})
			if !ok {
				// Value can't be pushed when the last value is not a list
				return fmt.Errorf("failed to push data to payload, final path element is not a list: \"%s\"", path)
			}
			// Finally make sure current points to the list with the added value
			current[name] = append(items, value)
			break
		}

		// Update the pointer to the current data element to match the path
		v, ok := current[name]
		if !ok {
			// When there is no value for current part create an empty map
			item := make(map[string]interface{})
			current[name] = item
			// Update current pointer to be the new item
			current = item
		} else if item, ok := v.(map[string]interface{}); ok {
			// Update pointer when the value is a map
			current = item
		} else {
			// The value is not a map, and because of that it can't be traversed
			return fmt.Errorf("failed to push data to payload, path is not traversable: \"%s\"", path)
		}
	}
	return nil
}

// IsError checks if payload is an error payload entity.
func (p Payload) IsError() bool {
	return p.Exists("error")
}

// IsCall checks if payload is a Service call payload entity.
func (p Payload) IsCall() bool {
	return p.Exists("call")
}

// IsCommand checks if payload is a command payload entity.
func (p Payload) IsCommand() bool {
	return p.Exists("command/command")
}

// IsCommandReply checks if payload is a command reply payload entity.
func (p Payload) IsCommandReply() bool {
	return p.Exists("command_reply")
}

// IsResponse checks if payload is a response payload entity.
// The semantic of the response itself can be HTTP or any other of the supported request types.
func (p Payload) IsResponse() bool {
	return p.Exists("response")
}

// IsTransport checks if payload is a transport payload entity.
func (p Payload) IsTransport() bool {
	return p.Exists("transport")
}

// Error gets an error payload initialized with current payload.
func (p Payload) Error() (*Error, bool) {
	if p.IsError() {
		return NewErrorFromMap(p.GetMap("error")), true
	}
	return nil, false
}

// Pack serializes the payload to a binary.
func (p Payload) Pack() ([]byte, error) {
	data, err := transform.Pack(p.Data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// Json serializes the payload to JSON.
func (p Payload) Json(pretty bool) ([]byte, error) {
	data, err := transform.Serialize(p.Data, pretty)
	if err != nil {
		return nil, err
	}
	return data, nil
}
