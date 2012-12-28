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

type map_ struct {
	length int32
	head   *mapentry
}

type mapentry struct {
	next *mapentry
	// after this comes the key, then the value.
}

// #llgo name: reflect.makemap
func reflect_makemap(t *map_) unsafe.Pointer {
	return makemap(unsafe.Pointer(t))
}

func makemap(t unsafe.Pointer) unsafe.Pointer {
	m := (*map_)(malloc(uintptr(unsafe.Sizeof(map_{}))))
	if m != nil {
		m.length = 0
		m.head = nil
	}
	return unsafe.Pointer(m)
}

// #llgo name: reflect.maplen
func reflect_maplen(m unsafe.Pointer) int32 {
	return int32(maplen((*map_)(m)))
}

func maplen(m *map_) int {
	if m != nil {
		return int(m.length)
	}
	return 0
}

// #llgo name: reflect.mapassign
func reflect_mapassign(t *rtype, m_, key, val unsafe.Pointer, ok bool) {
	m := (*map_)(m_)
	if ok {
		ptr := maplookup(t, m, key, true)
		// TODO use copy alg
		memmove(ptr, val, t.size)
	} else {
		mapdelete(t, m, key)
	}
}

// #llgo name: reflect.mapaccess
func reflect_mapaccess(t *rtype, m_, key unsafe.Pointer) (val unsafe.Pointer, ok bool) {
	m := (*map_)(m_)
	ptr := maplookup(t, m, key, false)
	return ptr, ptr != nil
}

func maplookup(t unsafe.Pointer, m *map_, key unsafe.Pointer, insert bool) unsafe.Pointer {
	if m == nil {
		return 0
	}

	maptyp := (*mapType)(unsafe.Pointer(t))
	ptrsize := uintptr(unsafe.Sizeof(m.head.next))
	keysize := uintptr(maptyp.key.size)
	keyoffset := align(ptrsize, uintptr(maptyp.key.align))
	elemsize := uintptr(maptyp.elem.size)
	elemoffset := align(keyoffset+keysize, uintptr(maptyp.elem.align))
	entrysize := elemoffset + elemsize

	// Search for the entry with the specified key.
	keyalgs := unsafe.Pointer(maptyp.key.alg)
	keyeqptr := unsafe.Pointer(uintptr(keyalgs) + unsafe.Sizeof(maptyp.key.alg))
	keyeqfun := *(*equalalg)(keyeqptr)
	var last *mapentry
	for ptr := m.head; ptr != nil; ptr = ptr.next {
		keyptr := unsafe.Pointer(uintptr(unsafe.Pointer(ptr)) + keyoffset)
		if keyeqfun(keysize, key, keyptr) {
			elemptr := unsafe.Pointer(uintptr(unsafe.Pointer(ptr)) + elemoffset)
			return elemptr
		}
		last = ptr
	}

	// Not found: insert the key if requested.
	if insert {
		newentry := (*mapentry)(malloc(entrysize))
		newentry.next = nil
		keyptr := unsafe.Pointer(uintptr(unsafe.Pointer(newentry) + keyoffset))
		elemptr := unsafe.Pointer(uintptr(unsafe.Pointer(newentry)) + elemoffset)
		memcpy(keyptr, key, keysize)
		if last != nil {
			last.next = newentry
		} else {
			m.head = newentry
		}
		m.length++
		return elemptr
	}

	return 0
}

func mapdelete(t unsafe.Pointer, m *map_, key unsafe.Pointer) {
	if m == nil {
		return 0
	}

	maptyp := (*mapType)(unsafe.Pointer(t))
	ptrsize := uintptr(unsafe.Sizeof(m.head.next))
	keysize := uintptr(maptyp.key.size)
	keyoffset := align(ptrsize, uintptr(maptyp.key.align))

	// Search for the entry with the specified key.
	keyalgs := unsafe.Pointer(maptyp.key.alg)
	keyeqptr := unsafe.Pointer(uintptr(keyalgs) + unsafe.Sizeof(maptyp.key.alg))
	keyeqfun := *(*equalalg)(keyeqptr)
	var last *mapentry
	for ptr := m.head; ptr != nil; ptr = ptr.next {
		keyptr := unsafe.Pointer(uintptr(unsafe.Pointer(ptr)) + keyoffset)
		if keyeqfun(keysize, key, keyptr) {
			if last == nil {
				m.head = ptr.next
			} else {
				last.next = ptr.next
			}
			free(ptr)
			m.length--
			return
		}
		last = ptr
	}
}

// #llgo name: reflect.mapiterinit
func reflect_mapiterinit(t *rtype, m_ unsafe.Pointer) *byte {
	// TODO
	return nil
}

// #llgo name: reflect.mapiterkey
func reflect_mapiterkey(it *byte) (key unsafe.Pointer, ok bool) {
	// TODO
	return
}

// #llgo name: reflect.mapiternext
func reflect_mapiternext(it *byte) {
	// TODO
}

func mapnext(t unsafe.Pointer, m *map_, nextin unsafe.Pointer) (nextout, pk, pv unsafe.Pointer) {
	if m == nil {
		return
	}
	ptr := (*mapentry)(nextin)
	if ptr == nil {
		ptr = m.head
	} else {
		ptr = ptr.next
	}
	if ptr != nil {
		maptyp := (*mapType)(t)
		ptrsize := uintptr(unsafe.Sizeof(m.head.next))
		keysize := uintptr(maptyp.key.size)
		keyoffset := align(ptrsize, uintptr(maptyp.key.align))
		elemsize := uintptr(maptyp.elem.size)
		elemoffset := align(keyoffset+keysize, uintptr(maptyp.elem.align))
		nextout = unsafe.Pointer(ptr)
		pk = unsafe.Pointer(uintptr(unsafe.Pointer(ptr)) + keyoffset)
		pv = unsafe.Pointer(uintptr(unsafe.Pointer(ptr)) + elemoffset)
	}
	return
}
