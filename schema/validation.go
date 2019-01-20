// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2019 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package schema

import (
	"fmt"

	jsonschema "github.com/xeipuuv/gojsonschema"
)

// Validate validates data using a JSON schema
func Validate(schema map[string]interface{}, data interface{}) error {
	s := jsonschema.NewGoLoader(schema)
	d := jsonschema.NewGoLoader(data)
	result, err := jsonschema.Validate(s, d)
	if err != nil {
		return err
	}

	if !result.Valid() {
		return fmt.Errorf("%s", result.Errors()[0])
	}

	return nil
}
