/*
Copyright (C)  2018 Yahoo Japan Corporation Athenz team.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package infra

import (
	"net/http/httputil"
	"sync"
	"sync/atomic"
)

type buffer struct {
	pool sync.Pool
	size *uint64
}

// NewBuffer implements httputil.BufferPool for providing byte slices of same size.
func NewBuffer(size uint64) httputil.BufferPool {
	if size == 0 {
		return nil
	}

	b := &buffer{
		size: &size,
	}

	b.pool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 0, atomic.LoadUint64(b.size))
		},
	}

	return b
}

// Get returns a slice from the pool, and remove it from the pool. New slice may be created when needed.
func (b *buffer) Get() []byte {
	return b.pool.Get().([]byte)
}

// Put adds the given slice back to internal pool, but resets its length to 0.
// If the given slice have capacity > current new buffer size in the pool, all newly created byte slices from Get() will have capacity which equals to the given slice. Capacity of slices already in the pool will not be affected.
func (b *buffer) Put(buf []byte) {
	size := atomic.LoadUint64(b.size)

	// The maximum capacity for a slice is the size of the default integer on the target build.
	// uint = either 32 or 64 bits; int = same size as uint
	// (some go implementation may have int > 64-bit?, although it is very rare to have response that big...)
	// convert int to int64 to prevent overflow
	bufLen := uint64(len(buf))
	bufCap := uint64(cap(buf))

	if bufLen >= size || bufCap >= size {
		// expand size if given buffer is larger than current size
		size = max(bufLen, bufCap) // len() <= cap() always true without unsafe

		// if len() <= cap() always true, may be no need to make new array
		buf = make([]byte, 0, size)
		atomic.StoreUint64(b.size, size)
	}

	// reset buffer length to 0
	b.pool.Put(buf[:0])
}

// max is copied from math.Max for uint64 type
func max(x, y uint64) uint64 {
	if x > y {
		return x
	}
	return y
}
