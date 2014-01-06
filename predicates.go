// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file implements commonly used type predicates.

package llgo

import (
	"code.google.com/p/go.tools/go/types"
)

func isBoolean(typ types.Type) bool {
	t, ok := typ.Underlying().(*types.Basic)
	return ok && t.Info()&types.IsBoolean != 0
}

func isInteger(typ types.Type) bool {
	t, ok := typ.Underlying().(*types.Basic)
	return ok && t.Info()&types.IsInteger != 0
}

func isUnsigned(typ types.Type) bool {
	t, ok := typ.Underlying().(*types.Basic)
	return ok && t.Info()&types.IsUnsigned != 0
}

func isFloat(typ types.Type) bool {
	t, ok := typ.Underlying().(*types.Basic)
	return ok && t.Info()&types.IsFloat != 0
}

func isComplex(typ types.Type) bool {
	t, ok := typ.Underlying().(*types.Basic)
	return ok && t.Info()&types.IsComplex != 0
}

func isString(typ types.Type) bool {
	t, ok := typ.Underlying().(*types.Basic)
	return ok && t.Info()&types.IsString != 0
}

func isUntyped(typ types.Type) bool {
	t, ok := typ.Underlying().(*types.Basic)
	return ok && t.Info()&types.IsUntyped != 0
}
