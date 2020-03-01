// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2020 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package payload

import (
	"errors"
	"fmt"
	"time"

	"github.com/kusanagi/kusanagi-sdk-go/format"
	"github.com/kusanagi/kusanagi-sdk-go/logging"
	"github.com/kusanagi/kusanagi-sdk-go/traverse"
	"github.com/kusanagi/kusanagi-sdk-go/version"
)

// GatewayAddr defines the addresses for a gateway.
type GatewayAddr struct {
	Internal string
	Public   string
}

type Fallback []interface{}

func (f Fallback) GetService() string {
	if len(f) == 0 {
		return ""
	}
	v, _ := f[0].(string)
	return v
}

func (f Fallback) GetVersion() string {
	if len(f) < 2 {
		return ""
	}
	v, _ := f[1].(string)
	return v
}

func (f Fallback) GetActions() []string {
	if len(f) < 3 {
		return nil
	}
	actions := make([]string, 0)
	if values, ok := f[2].([]interface{}); ok {
		for _, a := range values {
			actions = append(actions, a.(string))
		}
	}
	return actions
}

// NewMeta creates a new request/response meta payload.
func NewMeta(version, id, protocol, client string, gateway *GatewayAddr, rdt string) *Meta {
	m := Meta{Payload: NewNamespaced("meta")}
	m.SetVersion(version)
	m.SetID(id)
	m.SetDatetime(rdt)
	m.SetProtocol(protocol)
	m.SetGateway(gateway)
	m.SetClient(client)
	return &m
}

// Meta defines a request/response meta payload.
type Meta struct {
	*Payload
}

// GetVersion gets the version of the framework.
func (m Meta) GetVersion() string {
	return m.GetDefault("version", version.Get()).(string)
}

// SetVersion sets the version of the framework.
func (m *Meta) SetVersion(value string) error {
	return m.Set("version", value)
}

// GetID gets the request UUID.
func (m Meta) GetID() string {
	return m.GetString("id")
}

// SetID sets the request UUID.
func (m *Meta) SetID(value string) error {
	return m.Set("id", value)
}

// GetDatetime gets the date and time of the request.
func (m Meta) GetDatetime() string {
	return m.GetString("datetime")
}

// SetDatetime sets the date and time of the request.
func (m *Meta) SetDatetime(value string) error {
	if value == "" {
		// When no time is defined use current time
		value = format.TimeToString(time.Now().UTC())
	}

	return m.Set("datetime", value)
}

// GetMiddlewareType gets the type of middleware.
func (m Meta) GetMiddlewareType() int {
	return m.GetInt("type")
}

// SetMiddlewareType sets the type of middleware.
func (m *Meta) SetMiddlewareType(value int) error {
	return m.Set("type", value)
}

// GetProtocol gets the protocol implemented by the Gateway.
func (m Meta) GetProtocol() string {
	return m.GetString("protocol")
}

// SetProtocol sets the protocol implemented by the Gateway.
func (m *Meta) SetProtocol(value string) error {
	return m.Set("protocol", value)
}

// GetGateway gets the internal and public Gateway addresses.
func (m Meta) GetGateway() *GatewayAddr {
	if value := m.GetDefault("gateway", nil); value != nil {
		s := value.([]string)
		return &GatewayAddr{s[0], s[1]}
	}
	return nil
}

// SetGateway sets the internal and public Gateway addresses.
func (m *Meta) SetGateway(g *GatewayAddr) error {
	return m.Set("gateway", []interface{}{g.Internal, g.Public})
}

// GetClient gets the address of the client that sent the request.
// Client is available only for response meta.
func (m Meta) GetClient() string {
	return m.GetString("client")
}

// SetClient sets the address of the client that sent the request.
func (m *Meta) SetClient(value string) error {
	return m.Set("client", value)
}

// SetAttributes sets the request attributes.
// Request attributes are setted by the request middleware and are
// ONLY accessed as read only by the response middleware.
// Services must not receive them.
func (m *Meta) SetAttributes(attrs map[string]interface{}) error {
	return m.Set("attributes", attrs)
}

// NewTransportMetaFromMap creates a new transport meta payload from a map.
func NewTransportMetaFromMap(m map[string]interface{}) *TransportMeta {
	tm := TransportMeta{Payload: NewNamespaced("meta")}
	tm.Data = m
	return &tm
}

// NewTransportMeta creates a new transport meta payload.
func NewTransportMeta(version, rid, datetime string, gateway *GatewayAddr, origin []string, level int64) *TransportMeta {
	tm := TransportMeta{Payload: NewNamespaced("meta")}
	tm.SetVersion(version)
	tm.SetID(rid)
	tm.SetDatetime(datetime)
	tm.SetGateway(gateway)
	tm.SetLevel(level)
	tm.SetOrigin(origin)
	return &tm
}

// TransportMeta defines a transport meta payload.
type TransportMeta struct {
	*Payload
}

// GetVersion gets the version of the framework.
func (tm TransportMeta) GetVersion() string {
	return tm.GetDefault("version", version.Get()).(string)
}

// SetVersion sets the version of the framework.
func (tm *TransportMeta) SetVersion(value string) error {
	return tm.Set("version", value)
}

// GetID gets the request UUID.
func (tm TransportMeta) GetID() string {
	return tm.GetString("id")
}

// SetID sets the request UUID.
func (tm *TransportMeta) SetID(value string) error {
	return tm.Set("id", value)
}

// GetLevel gets the request level.
func (tm TransportMeta) GetLevel() int64 {
	return tm.GetInt64("level")
}

// SetLevel sets the request level.
func (tm *TransportMeta) SetLevel(value int64) error {
	if value < 1 {
		return errors.New("transport level must be greater than 0")
	}
	return tm.Set("level", value)
}

// GetDatetime gets the date and time of the request.
func (tm TransportMeta) GetDatetime() string {
	return tm.GetString("datetime")
}

// SetDatetime sets the date and time of the request.
func (tm *TransportMeta) SetDatetime(value string) error {
	// When no time is defined use current time
	if value == "" {
		value = format.TimeToString(time.Now().UTC())
	}
	return tm.Set("datetime", value)
}

// GetGateway gets the internal and public Gateway addresses.
func (tm TransportMeta) GetGateway() *GatewayAddr {
	if values := tm.GetSlice("gateway"); values != nil {
		return &GatewayAddr{values[0].(string), values[1].(string)}
	}
	return nil
}

// SetGateway sets the internal and public Gateway addresses.
func (tm *TransportMeta) SetGateway(g *GatewayAddr) error {
	return tm.Set("gateway", []interface{}{g.Internal, g.Public})
}

// GetOrigin gets the origin Service for the request.
func (tm TransportMeta) GetOrigin() []string {
	if values := tm.GetSlice("origin"); values != nil {
		res := []string{}
		for _, v := range values {
			res = append(res, v.(string))
		}
		return res
	}
	return nil
}

// SetOrigin sets the origin Service for the request.
func (tm *TransportMeta) SetOrigin(value []string) error {
	origin := []interface{}{}
	for _, v := range value {
		origin = append(origin, v)
	}
	return tm.Set("origin", origin)
}

// GetFallbacks gets triggered Service fallbacks.
func (tm TransportMeta) GetFallbacks() []Fallback {
	if values := tm.GetSlice("fallbacks"); len(values) > 0 {
		fallbacks := make([]Fallback, 0)
		for _, v := range values {
			if f, ok := v.([]interface{}); ok && len(f) > 0 {
				fallbacks = append(fallbacks, Fallback(f))
			}
		}
		return fallbacks
	}
	return nil
}

// SetFallbacks sets triggered Service fallbacks.
func (tm *TransportMeta) SetFallbacks(value [][]interface{}) error {
	return tm.Set("fallbacks", value)
}

// GetProperties gets custom userland properties.
func (tm TransportMeta) GetProperties() map[string]string {
	if v := tm.GetMap("properties"); v != nil {
		p := make(map[string]string)
		for name, value := range v {
			p[name], _ = value.(string)
		}
		return p
	}
	return nil
}

// SetProperties sets custom userland properties.
func (tm *TransportMeta) SetProperties(value map[string]string) error {
	return tm.Set("properties", value)
}

// SetProperty sets a custom userland property.
func (tm *TransportMeta) SetProperty(name, value string) error {
	return tm.Set(fmt.Sprintf("properties/%s", traverse.N(name)), value)
}

// GetStartTime gets the time the request started.
func (tm TransportMeta) GetStartTime() *time.Time {
	if s := tm.GetString("start_time"); s != "" {
		t, err := format.TimeFromString(s)
		if err != nil {
			logging.Error("invalid start time format in transport meta")
			return nil
		}
		return &t
	}
	return nil
}

// SetStartTime sets the time the request started.
func (tm *TransportMeta) SetStartTime(t time.Time) error {
	return tm.Set("start_time", format.TimeToString(t))
}

// GetEndTime gets the time the request finished.
func (tm TransportMeta) GetEndTime() *time.Time {
	if s := tm.GetString("end_time"); s != "" {
		t, err := format.TimeFromString(s)
		if err != nil {
			logging.Error("invalid end time format in transport meta")
			return nil
		}
		return &t
	}
	return nil
}

// SetEndTime sets the time the request finished.
func (tm *TransportMeta) SetEndTime(t time.Time) error {
	return tm.Set("end_time", format.TimeToString(t))
}

// GetDuration gets duration time for the request.
func (tm TransportMeta) GetDuration() int64 {
	return tm.GetInt64("duration")
}

// SetDuration sets duration time for the request.
func (tm *TransportMeta) SetDuration(d int64) error {
	return tm.Set("duration", d)
}
