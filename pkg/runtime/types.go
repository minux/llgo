/*
Copyright (c) 2011, 2012 Andrew Wilkins <axwalk@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
of the Software, and to permit persons to whom the Software is furnished to do
so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package runtime

import "unsafe"

// These types are based on those from go/src/pkg/reflect/type.go, must
// keep in sync!

// rtype is the common implementation of most values.
// It is embedded in other, public struct types, but always
// with a unique tag like `reflect:"array"` or `reflect:"ptr"`
// so that code cannot convert from, say, *arrayType to *ptrType.
type rtype struct {
	size          uintptr  // size in bytes
	hash          uint32   // hash of type; avoids computation in hash tables
	_             uint8    // unused/padding
	align         uint8    // alignment of variable with this type
	fieldAlign    uint8    // alignment of struct field with this type
	kind          uint8    // enumeration for C
	alg           *uintptr // algorithm table (../runtime/runtime.h:/Alg)
	gc            uintptr  // garbage collection data
	string        *string  // string form; unnecessary but undeniably useful
	*uncommonType          // (relatively) uncommon fields
	ptrToThis     *rtype   // type for pointer to this type, if used in binary or has methods
}

// Method on non-interface type
type method struct {
	name    *string        // name of method
	pkgPath *string        // nil for exported Names; otherwise import path
	mtyp    *rtype         // method type (without receiver)
	typ     *rtype         // .(*FuncType) underneath (with receiver)
	ifn     unsafe.Pointer // fn used in interface call (one-word receiver)
	tfn     unsafe.Pointer // fn used for normal method call
}

// uncommonType is present only for types with names or methods
// (if T is a named type, the uncommonTypes for T and *T have methods).
// Using a pointer to this struct reduces the overall size required
// to describe an unnamed type with no methods.
type uncommonType struct {
	name    *string  // name of type
	pkgPath *string  // import path; nil for built-in types like int, string
	methods []method // methods associated with type
}

// arrayType represents a fixed array type.
type arrayType struct {
	rtype `reflect:"array"`
	elem  *rtype // array element type
	slice *rtype // slice type
	len   uintptr
}

// chanType represents a channel type.
type chanType struct {
	rtype `reflect:"chan"`
	elem  *rtype  // channel element type
	dir   uintptr // channel direction (ChanDir)
}

// funcType represents a function type.
type funcType struct {
	rtype     `reflect:"func"`
	dotdotdot bool     // last input parameter is ...
	in        []*rtype // input parameter types
	out       []*rtype // output parameter types
}

// imethod represents a method on an interface type
type imethod struct {
	name    *string // name of method
	pkgPath *string // nil for exported Names; otherwise import path
	typ     *rtype  // .(*FuncType) underneath
}

// interfaceType represents an interface type.
type interfaceType struct {
	rtype   `reflect:"interface"`
	methods []imethod // sorted by hash
}

// mapType represents a map type.
type mapType struct {
	rtype `reflect:"map"`
	key   *rtype // map key type
	elem  *rtype // map element (value) type
}

// ptrType represents a pointer type.
type ptrType struct {
	rtype `reflect:"ptr"`
	elem  *rtype // pointer element (pointed at) type
}

// sliceType represents a slice type.
type sliceType struct {
	rtype `reflect:"slice"`
	elem  *rtype // slice element type
}

// Struct field
type structField struct {
	name    *string // nil for embedded fields
	pkgPath *string // nil for exported Names; otherwise import path
	typ     *rtype  // type of field
	tag     *string // nil if no tag
	offset  uintptr // byte offset of field within struct
}

// structType represents a struct type.
type structType struct {
	rtype  `reflect:"struct"`
	fields []structField // sorted by offset
}

const (
	invalidKind uint8 = iota
	boolKind
	intKind
	int8Kind
	int16Kind
	int32Kind
	int64Kind
	uintKind
	uint8Kind
	uint16Kind
	uint32Kind
	uint64Kind
	uintptrKind
	float32Kind
	float64Kind
	complex64Kind
	complex128Kind
	arrayKind
	chanKind
	funcKind
	interfaceKind
	mapKind
	ptrKind
	sliceKind
	stringKind
	structKind
	unsafePointerKind
)

// eqtyp takes two runtime types and returns true
// iff they are equal.
func eqtyp(t1, t2 *rtype) bool {
	if t1 == t2 {
		return true
	}
	if t1.kind == t2.kind {
		// TODO check rules for type equality.
		//
		// Named type equality is covered in the trivial
		// case, since there is only one definition for
		// each named type.
		// 
		// Basic types are not checked for explicitly,
		// as we should never be comparing unnamed basic
		// types.
		switch t1.kind {
		case arrayKind:
		case chanKind:
		case funcKind:
		case interfaceKind:
		case mapKind:
		case ptrKind:
			t1 := (*ptrType)(unsafe.Pointer(t1))
			t2 := (*ptrType)(unsafe.Pointer(t2))
			return eqtyp(t1.elem, t2.elem)
		case sliceKind:
		case structKind:
		}
	}
	return false
}

///////////////////////////////////////////////////////////////////////////////
// Types used in runtime function signatures.

type _string struct {
	str *uint8
	len int
}

type slice struct {
	array *uint8
	len   uint
	cap   uint
}
