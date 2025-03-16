package id3v2

import (
	"bufio"
	"io"
)

// bufferedWriter is a utility struct for writing ID3v2 frames efficiently.
// It wraps a bufio.Writer and tracks the number of bytes written, while also
// handling errors gracefully to avoid unnecessary writes after an error occurs.
type bufferedWriter struct {
	err     error         // Stores the first error encountered during writing.
	w       *bufio.Writer // Underlying buffered writer for efficient I/O.
	written int           // Tracks the total number of bytes written so far.
}

// newBufferedWriter initializes a new bufferedWriter with the provided io.Writer.
// It wraps the writer in a bufio.Writer for buffered I/O operations.
func newBufferedWriter(w io.Writer) *bufferedWriter {
	return &bufferedWriter{w: bufio.NewWriter(w)}
}

// EncodeAndWriteText encodes the provided string using the specified encoding
// and writes it to the underlying writer. If an error occurs during encoding
// or writing, it is stored in the bufferedWriter's err field.
func (bw *bufferedWriter) EncodeAndWriteText(src string, to Encoding) {
	if bw.err != nil {
		return // Skip if an error has already occurred.
	}

	bw.err = encodeWriteText(bw, src, to) // Encode and write the text.
}

// Flush flushes any buffered data to the underlying writer.
// If an error occurred during any previous write operation, it is returned.
func (bw *bufferedWriter) Flush() error {
	if bw.err != nil {
		return bw.err // Return the first error encountered.
	}

	return bw.w.Flush() // Flush the buffered writer.
}

// Reset resets the bufferedWriter to use a new io.Writer.
// It also clears any stored errors and resets the byte counter.
func (bw *bufferedWriter) Reset(w io.Writer) {
	bw.err = nil
	bw.written = 0

	bw.w.Reset(w) // Reset the underlying bufio.Writer.
}

// WriteByte writes a single byte to the underlying writer.
// If an error occurs, it is stored in the bufferedWriter's err field.
//
//nolint:govet // This method does not implement io.ByteWriter.
func (bw *bufferedWriter) WriteByte(c byte) {
	if bw.err != nil {
		return // Skip if an error has already occurred.
	}

	bw.err = bw.w.WriteByte(c) // Write the byte.
	if bw.err == nil {
		bw.written++ // Increment the byte counter if successful.
	}
}

// WriteBytesSize writes the size in bytes to the underlying writer.
// The size is written in either synch-safe or non-synch-safe format,
// depending on the synchSafe parameter.
func (bw *bufferedWriter) WriteBytesSize(size uint, synchSafe bool) {
	if bw.err != nil {
		return // Skip if an error has already occurred.
	}

	bw.err = writeBytesSize(bw, size, synchSafe) // Write the size.
}

// WriteString writes a string to the underlying writer.
// If an error occurs, it is stored in the bufferedWriter's err field.
func (bw *bufferedWriter) WriteString(s string) {
	if bw.err != nil {
		return // Skip if an error has already occurred.
	}

	var n int
	n, bw.err = bw.w.WriteString(s) // Write the string.
	bw.written += n                 // Increment the byte counter.
}

// Write writes a byte slice to the underlying writer.
// If an error occurs, it is stored in the bufferedWriter's err field.
func (bw *bufferedWriter) Write(p []byte) (n int, err error) {
	if bw.err != nil {
		return 0, bw.err // Skip if an error has already occurred.
	}

	n, err = bw.w.Write(p) // Write the byte slice.
	bw.written += n        // Increment the byte counter.
	bw.err = err           // Store any error that occurred.

	return n, err
}

// Written returns the total number of bytes written so far.
func (bw *bufferedWriter) Written() int {
	return bw.written
}

// useBufferedWriter is a helper function that simplifies the use of bufferedWriter.
// It ensures that the bufferedWriter is properly initialized and cleaned up.
// If the provided writer is already a bufferedWriter, it is reused.
// Otherwise, a new bufferedWriter is created and later returned to the pool.
func useBufferedWriter(w io.Writer, writeFunc func(*bufferedWriter) error) (int64, error) {
	var initialBytesWritten int

	// Check if the writer is already a bufferedWriter.
	bufferedWriter, isBuffered := w.(*bufferedWriter)
	if isBuffered {
		initialBytesWritten = bufferedWriter.Written()
	} else {
		// Create a new bufferedWriter from the pool.
		bufferedWriter = getBufWriter(w)
		defer putBufWriter(bufferedWriter) // Return the bufferedWriter to the pool when done.
	}

	// Execute the provided write function.
	err := writeFunc(bufferedWriter)
	if err != nil {
		// Return the number of bytes written before the error occurred.
		return int64(initialBytesWritten), err
	}

	// Calculate the total number of bytes written during this operation.
	bytesWritten := int64(bufferedWriter.Written() - initialBytesWritten)

	// Flush any remaining data and return the result.
	if err = bufferedWriter.Flush(); err != nil {
		return bytesWritten, err
	}

	return bytesWritten, nil
}
