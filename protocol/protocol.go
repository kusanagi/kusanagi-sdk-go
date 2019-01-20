// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2019 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package protocol

// URN defines KUSANAGI protocol URNs.
var URN = map[string]string{
	"http": "urn:kusanagi:protocol:http",
	"ktp":  "urn:kusanagi:protocol:ktp",
}

// Meta values for the multipart responses.
// These values describe the response transport features.
const (
	EmptyMeta        = '\x00'
	ServiceCallMeta  = '\x01'
	FilesMeta        = '\x02'
	TransactionsMeta = '\x03'
	DownloadMeta     = '\x04'
)
