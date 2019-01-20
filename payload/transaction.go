// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2019 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package payload

// NewPayloadFromMap creates a new transactions payload from a map.
func NewTransactionsFromMap(data map[string]interface{}) *Transactions {
	return &Transactions{Payload: NewFromMap("", data)}
}

// New creates a new transactions payload.
func NewTransactions() *Transactions {
	return NewTransactionsFromMap(nil)
}

// Transactions represents a transactions payload.
// Transactions payload is part of the transport payload, and it contains
// the transaction actions to call grouped by action name.
type Transactions struct {
	*Payload
}

// HasActions checks if payload contains transaction actions for a command.
func (p Transactions) HasActions(command string) bool {
	return len(p.GetSlice(command)) > 0
}

// GetActions gets service actions for a transaction command.
func (p Transactions) GetActions(command string) (actions []*TransactionAction) {
	if !p.HasActions(command) {
		return nil
	}

	for _, v := range p.GetSlice(command) {
		// Get action payload data
		data, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		// Add the service action to call for current command
		actions = append(actions, NewTransactionActionFromMap(data))
	}
	return actions
}

// NewTransactionActionFromMap creates a new transaction action payload from a map.
func NewTransactionActionFromMap(data map[string]interface{}) *TransactionAction {
	return &TransactionAction{Payload: NewFromMap("", data)}
}

// New creates a new transaction action payload.
func NewTransactionAction() *TransactionAction {
	return NewTransactionActionFromMap(nil)
}

// TransactionAction represents a single transaction action.
type TransactionAction struct {
	*Payload
}

func (a TransactionAction) Name() string {
	return a.GetString("name")
}

func (a TransactionAction) Version() string {
	return a.GetString("version")
}

func (a TransactionAction) Action() string {
	return a.GetString("action")
}

func (a TransactionAction) Caller() string {
	return a.GetString("caller")
}

func (a TransactionAction) HasParams() bool {
	return a.Exists("params")
}

func (a TransactionAction) Params() (params []*Param) {
	if !a.HasParams() {
		return nil
	}

	for _, v := range a.GetSlice("params") {
		data, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		params = append(params, NewParamFromMap(data))
	}
	return params
}
