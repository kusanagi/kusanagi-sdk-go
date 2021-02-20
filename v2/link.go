// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2021 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package kusanagi

// Link represents a service link.
type Link struct {
	address   string
	service   string
	reference string
	uri       string
}

// GetAddress returns the gateway address for the service.
func (l Link) GetAddress() string {
	return l.address
}

// GetName returns the service name.
func (l Link) GetName() string {
	return l.service
}

// GetLink returns the link reference.
func (l Link) GetLink() string {
	return l.reference
}

// GetURI returns the link URI.
func (l Link) GetURI() string {
	return l.uri
}
