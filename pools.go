package id3v2

import (
	"bytes"
	"io"
	"sync"
)

// bsPool stores *[]byte. SA6002: a slice value in sync.Pool boxes the
// pointer/len/cap header on every Put; a pointer to the slice does not.
// See BenchmarkByteSlicePool32K and BenchmarkByteSlicePool128K in pools_test.go.
var bsPool = sync.Pool{
	New: func() any {
		b := make([]byte, 0)
		return &b
	},
}

// getByteSlice returns a byte slice of the specified size.
// It first tries to reuse a slice from the pool. If none is available, it allocates a new one.
func getByteSlice(size int) []byte {
	fromPool, _ := bsPool.Get().(*[]byte)
	if fromPool == nil {
		return make([]byte, size)
	}

	bs := *fromPool
	if cap(bs) < size {
		return make([]byte, size)
	}

	return bs[0:size]
}

// putByteSlice returns a byte slice to the pool for reuse.
func putByteSlice(b []byte) {
	if cap(b) == 0 {
		return
	}

	bsPool.Put(&b)
}

// bwPool reuses the internal bufio.Writer buffer.
// See BenchmarkBufferedWriterPool in pools_test.go.
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
	bw.Reset(nil)
	bwPool.Put(bw)
}

// rdPool reuses the internal bufio.Reader buffer.
// See BenchmarkBufferedReaderPool in pools_test.go.
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
	rd.Reset(nil)
	rdPool.Put(rd)
}

// bbPool reuses bytes.Buffer instances.
// See BenchmarkBytesBufferPool in pools_test.go.
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
