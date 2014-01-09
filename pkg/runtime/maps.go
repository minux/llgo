// Copyright 2011 The llgo Authors.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package runtime

import "unsafe"

type map_ []*mapentry

type mapentry struct {
	// first comes the key, then the value.
}

// #llgo name: reflect.ismapkey
func reflect_ismapkey(t *rtype) bool {
	// TODO
	return false
}

// #llgo name: reflect.makemap
func reflect_makemap(t *map_) unsafe.Pointer {
	return makemap(unsafe.Pointer(t), 0)
}

func makemap(t unsafe.Pointer, cap int) unsafe.Pointer {
	m := (*map_)(malloc(uintptr(unsafe.Sizeof(map_{}))))
	return unsafe.Pointer(m)
}

// #llgo name: reflect.maplen
func reflect_maplen(m unsafe.Pointer) int32 {
	return int32(maplen(m))
}

func maplen(m unsafe.Pointer) int {
	if m != nil {
		return len(*(*map_)(m))
	}
	return 0
}

// #llgo name: reflect.mapassign
func reflect_mapassign(t *mapType, m, key, val unsafe.Pointer, ok bool) {
	if ok {
		ptr := maplookup(unsafe.Pointer(t), m, key, true)
		// TODO use copy alg
		memmove(ptr, val, t.elem.size)
	} else {
		mapdelete(unsafe.Pointer(t), m, key)
	}
}

// #llgo name: reflect.mapaccess
func reflect_mapaccess(t *rtype, m, key unsafe.Pointer) (val unsafe.Pointer, ok bool) {
	ptr := maplookup(unsafe.Pointer(t), m, key, false)
	if ptr == nil {
		return nil, false
	}
	maptyp := (*mapType)(unsafe.Pointer(t))
	elemsize := uintptr(maptyp.elem.size)
	if elemsize <= unsafe.Sizeof(key) {
		memcpy(unsafe.Pointer(&val), ptr, elemsize)
	} else {
		val = ptr
	}
	return val, true
}

// mapaccess copies the value for the given key out if that key exists,
// else nil. The return value is true iff the key exists.
func mapaccess(t unsafe.Pointer, m_, key, outval unsafe.Pointer) bool {
	maptyp := (*mapType)(t)
	elemsize := uintptr(maptyp.elem.size)
	ptr := maplookup(t, m_, key, false)
	if ptr != nil {
		memcpy(outval, ptr, elemsize)
		return true
	}
	bzero(outval, elemsize)
	return false
}

// maplookup returns a pointer to the value for the given key
func maplookup(t unsafe.Pointer, m_, key unsafe.Pointer, insert bool) unsafe.Pointer {
	if m_ == nil {
		return nil
	}
	m := (*map_)(m_)

	maptyp := (*mapType)(t)
	ptrsize := uintptr(unsafe.Sizeof(m_))
	keysize := uintptr(maptyp.key.size)
	keyoffset := align(ptrsize, uintptr(maptyp.key.align))
	elemsize := uintptr(maptyp.elem.size)
	elemoffset := align(keyoffset+keysize, uintptr(maptyp.elem.align))
	entrysize := elemoffset + elemsize

	// Search for the entry with the specified key.
	keyalgs := unsafe.Pointer(maptyp.key.alg)
	keyeqptr := unsafe.Pointer(uintptr(keyalgs) + unsafe.Sizeof(maptyp.key.alg))
	keyeqfun := *(*unsafe.Pointer)(keyeqptr)
	for i := 0; i < len(*m); i++ {
		ptr := (*m)[i]
		keyptr := unsafe.Pointer(uintptr(unsafe.Pointer(ptr)) + keyoffset)
		if eqalg(keyeqfun, keysize, key, keyptr) {
			elemptr := unsafe.Pointer(uintptr(unsafe.Pointer(ptr)) + elemoffset)
			return elemptr
		}
	}

	// Not found: insert the key if requested.
	if insert {
		newentry := (*mapentry)(malloc(entrysize))
		keyptr := unsafe.Pointer(uintptr(unsafe.Pointer(newentry)) + keyoffset)
		elemptr := unsafe.Pointer(uintptr(unsafe.Pointer(newentry)) + elemoffset)
		memcpy(keyptr, key, keysize)
		*m = append(*m, newentry)
		return elemptr
	}

	return nil
}

func mapdelete(t unsafe.Pointer, m_, key unsafe.Pointer) {
	if m_ == nil {
		return
	}
	m := (*map_)(m_)

	maptyp := (*mapType)(t)
	ptrsize := uintptr(unsafe.Sizeof(m_))
	keysize := uintptr(maptyp.key.size)
	keyoffset := align(ptrsize, uintptr(maptyp.key.align))

	// Search for the entry with the specified key.
	keyalgs := unsafe.Pointer(maptyp.key.alg)
	keyeqptr := unsafe.Pointer(uintptr(keyalgs) + unsafe.Sizeof(maptyp.key.alg))
	keyeqfun := *(*unsafe.Pointer)(keyeqptr)
	for i := 0; i < len(*m); i++ {
		ptr := (*m)[i]
		keyptr := unsafe.Pointer(uintptr(unsafe.Pointer(ptr)) + keyoffset)
		if eqalg(keyeqfun, keysize, key, keyptr) {
			var tail []*mapentry
			if len(*m) > i+1 {
				tail = (*m)[i+1:]
			}
			(*m) = append((*m)[:i], tail...)
			free(unsafe.Pointer(ptr))
			return
		}
	}
}

// #llgo name: reflect.mapiterinit
func reflect_mapiterinit(t *rtype, m unsafe.Pointer) *byte {
	return (*byte)(mapiterinit(unsafe.Pointer(t), m))
}

// #llgo name: reflect.mapiterkey
func reflect_mapiterkey(iter_ *byte) (key unsafe.Pointer, ok bool) {
	if iter_ == nil {
		return nil, false
	}
	iter := (*mapiter)(unsafe.Pointer(iter_))
	if iter.i >= len(*iter.m) {
		return nil, false
	}
	entry := (*iter.m)[iter.i]
	keysize := uintptr(iter.typ.key.size)
	ptrsize := uintptr(unsafe.Sizeof(entry))
	keyoffset := align(ptrsize, uintptr(iter.typ.key.align))
	ptrk := unsafe.Pointer(uintptr(unsafe.Pointer(entry)) + keyoffset)
	if keysize <= unsafe.Sizeof(key) {
		memcpy(unsafe.Pointer(&key), ptrk, keysize)
	} else {
		key = ptrk
	}
	return key, true
}

// #llgo name: reflect.mapiternext
func reflect_mapiternext(iter_ *byte) {
	iter := (*mapiter)(unsafe.Pointer(iter_))
	iter.i++
}

type mapiter struct {
	typ *mapType
	m   *map_
	i   int
}

// TODO pass pointer to stack allocated block in
func mapiterinit(t, m unsafe.Pointer) unsafe.Pointer {
	if m == nil {
		return nil
	}
	iter := (*mapiter)(malloc(unsafe.Sizeof(mapiter{})))
	iter.typ = (*mapType)(t)
	iter.m = (*map_)(m)
	return unsafe.Pointer(iter)
}

func mapiternext(iter_, pk, pv unsafe.Pointer) bool {
	if iter_ == nil {
		return false
	}
	iter := (*mapiter)(iter_)
	keysize := uintptr(iter.typ.key.size)
	elemsize := uintptr(iter.typ.elem.size)
	if iter.i >= len(*iter.m) {
		bzero(pk, keysize)
		bzero(pv, elemsize)
		return false
	}
	entry := (*iter.m)[iter.i]
	iter.i++
	ptrsize := uintptr(unsafe.Sizeof(entry))
	keyoffset := align(ptrsize, uintptr(iter.typ.key.align))
	elemoffset := align(keyoffset+keysize, uintptr(iter.typ.elem.align))
	memcpy(pk, unsafe.Pointer(uintptr(unsafe.Pointer(entry))+keyoffset), keysize)
	memcpy(pv, unsafe.Pointer(uintptr(unsafe.Pointer(entry))+elemoffset), elemsize)
	return true
}
