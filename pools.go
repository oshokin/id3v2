package id3v2

import (
	"bytes"
	"io"
	"sync"
)

// bsPool is a pool of byte slices used to reduce allocations and improve performance.
// It stores reusable byte slices to avoid repeatedly allocating and freeing memory.
var bsPool = sync.Pool{
	New: func() any { return nil }, // If the pool is empty, return nil.
}

// getByteSlice returns a byte slice of the specified size.
// It first tries to reuse a slice from the pool. If none is available, it allocates a new one.
func getByteSlice(size int) []byte {
	fromPool := bsPool.Get()
	if fromPool == nil {
		return make([]byte, size) // Allocate a new slice if the pool is empty.
	}

	bs, _ := fromPool.([]byte)
	if cap(bs) < size {
		bs = make([]byte, size) // Allocate a new slice if the pooled slice is too small.
	}

	return bs[0:size] // Return a slice of the requested size.
}

// putByteSlice returns a byte slice to the pool for reuse.
// This helps reduce memory allocations by recycling slices.
func putByteSlice(b []byte) {
	//nolint:staticcheck // slice is already a pointer
	bsPool.Put(b) // Add the slice back to the pool.
}

// bwPool is a pool of buffered writers used to reduce allocations.
var bwPool = sync.Pool{
	New: func() any { return newBufferedWriter(nil) }, // Create a new bufferedWriter if the pool is empty.
}

// getBufWriter retrieves a buffered writer from the pool and resets it for reuse.
func getBufWriter(w io.Writer) *bufferedWriter {
	bw, _ := bwPool.Get().(*bufferedWriter)
	bw.Reset(w) // Reset the writer to the new io.Writer.

	return bw
}

// putBufWriter returns a buffered writer to the pool for reuse.
func putBufWriter(bw *bufferedWriter) {
	bwPool.Put(bw) // Add the writer back to the pool.
}

// lrPool is a pool of io.LimitedReader instances used to reduce allocations.
var lrPool = sync.Pool{
	New: func() any { return new(io.LimitedReader) }, // Create a new LimitedReader if the pool is empty.
}

// getLimitedReader retrieves a LimitedReader from the pool and initializes it with the given reader and limit.
func getLimitedReader(rd io.Reader, n int64) *io.LimitedReader {
	r, _ := lrPool.Get().(*io.LimitedReader)
	r.R = rd // Set the underlying reader.
	r.N = n  // Set the read limit.

	return r
}

// putLimitedReader returns a LimitedReader to the pool for reuse.
func putLimitedReader(r *io.LimitedReader) {
	r.N = 0       // Reset the read limit.
	r.R = nil     // Clear the underlying reader.
	lrPool.Put(r) // Add the reader back to the pool.
}

// rdPool is a pool of buffered readers used to reduce allocations.
var rdPool = sync.Pool{
	New: func() any { return newBufferedReader(nil) }, // Create a new bufferedReader if the pool is empty.
}

// getBufReader retrieves a buffered reader from the pool and resets it for reuse.
func getBufReader(rd io.Reader) *bufferedReader {
	reader, _ := rdPool.Get().(*bufferedReader)
	reader.Reset(rd) // Reset the reader to the new io.Reader.

	return reader
}

// putBufReader returns a buffered reader to the pool for reuse.
func putBufReader(rd *bufferedReader) {
	rdPool.Put(rd) // Add the reader back to the pool.
}

// bbPool is a pool of bytes.Buffer instances used to reduce allocations.
var bbPool = sync.Pool{
	New: func() any { return new(bytes.Buffer) }, // Create a new bytes.Buffer if the pool is empty.
}

// getBytesBuffer retrieves a bytes.Buffer from the pool.
func getBytesBuffer() *bytes.Buffer {
	result, _ := bbPool.Get().(*bytes.Buffer)

	return result
}

// putBytesBuffer returns a bytes.Buffer to the pool for reuse.
func putBytesBuffer(buf *bytes.Buffer) {
	buf.Reset()     // Clear the buffer's contents.
	bbPool.Put(buf) // Add the buffer back to the pool.
}
