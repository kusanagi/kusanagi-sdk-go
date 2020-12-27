// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2020 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package kusanagi

import "github.com/kusanagi/kusanagi-sdk-go/v2/lib/payload"

// Transaction commands
const Commit = payload.TransactionCommit
const Rollback = payload.TransactionRollback
const Complete = payload.TransactionComplete

// Transaction represents a single transaction.
type Transaction struct {
	command string
	name    string
	version string
	action  string
	caller  string
	params  []*Param
}

// GetType returns the transaction command type.
func (t Transaction) GetType() string {
	return t.command
}

// GetVersion returns the name of the service that registered the transaction.
func (t Transaction) GetName() string {
	return t.name
}

// GetVersion returns the version of the service that registered the transaction.
func (t Transaction) GetVersion() string {
	return t.version
}

// GetCallerAction returns the name of the action that registered the transaction.
func (t Transaction) GetCallerAction() string {
	return t.caller
}

// GetCalleeAction returns the name of the action to be called by the transaction.
func (t Transaction) GetCalleeAction() string {
	return t.action
}

// GetParams gets the transaction parameters.
func (t Transaction) GetParams() (params []*Param) {
	// Add the parameters to a new list
	for _, p := range t.params {
		params = append(params, p)
	}
	return params
}
