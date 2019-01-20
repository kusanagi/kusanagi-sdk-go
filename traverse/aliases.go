// Go SDK for the KUSANAGI(tm) framework (http://kusanagi.io)
// Copyright (c) 2016-2019 KUSANAGI S.L. All rights reserved.
//
// Distributed under the MIT license.
//
// For the full copyright and license information, please view the LICENSE
// file that was distributed with this source code.

package traverse

var ignoreAlias = "!"

// N marks a path element so its name is not changed to an alias.
func N(name string) string {
	return ignoreAlias + name
}

// Aliases defines a type to save field name aliases.
type Aliases map[string]string

// Get gets an alias for a name.
func (a Aliases) Get(name string) string {
	// When name has the ignore alias prefix return the name without the prefix
	if string(name[0]) == ignoreAlias {
		return name[1:]
	}

	// When alias exists return it, otherwise return the original name
	if alias, ok := a[name]; ok {
		return alias
	}
	return name
}

// Update updates the aliases with values from other aliase object.
func (a Aliases) Update(m Aliases) Aliases {
	for k, v := range m {
		a[k] = v
	}
	return a
}
