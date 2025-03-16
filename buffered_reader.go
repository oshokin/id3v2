package id3v2

import (
	"bufio"
	"bytes"
	"io"
)

// bufferedReader is a utility for conveniently parsing ID3v2 frames.
// It wraps a bufio.Reader and tracks errors encountered during reading.
type bufferedReader struct {
	buf *bufio.Reader // The underlying buffered reader.
	err error         // Stores the last error encountered during reading.
}

// newBufferedReader creates and returns a new bufferedReader instance
// initialized with the provided io.Reader.
func newBufferedReader(rd io.Reader) *bufferedReader {
	return &bufferedReader{buf: bufio.NewReader(rd)}
}

// Discard skips the next n bytes in the buffer.
// If an error has already occurred, it does nothing.
func (br *bufferedReader) Discard(n int) {
	if br.err != nil {
		return
	}

	_, br.err = br.buf.Discard(n)
}

// Err returns the last error encountered during reading.
func (br *bufferedReader) Err() error {
	return br.err
}

// Read reads data into p using the underlying bufio.Reader.
// If an error has already occurred, it does nothing and returns the error.
// Note: Errors from br.buf.Read(p) are not stored in br.err.
func (br *bufferedReader) Read(p []byte) (n int, err error) {
	if br.err != nil {
		return 0, br.err
	}

	return br.buf.Read(p)
}

// ReadAll reads all remaining data from the buffer until an error or EOF.
// It returns the data as a byte slice. A successful call returns err == nil,
// even if EOF is encountered.
func (br *bufferedReader) ReadAll() []byte {
	if br.err != nil {
		return nil
	}

	// Create a buffer to store the read data.
	buf := bytes.NewBuffer(make([]byte, 0, bytes.MinRead))

	// Read all data from the bufferedReader into the buffer.
	_, err := buf.ReadFrom(br)
	if err != nil && br.err == nil {
		br.err = err // Store the error if one occurs.

		return nil
	}

	return buf.Bytes()
}

// ReadByte reads and returns a single byte from the buffer.
// If an error has already occurred, it returns 0.
//
//nolint:govet // This method does not implement io.ByteReader.
func (br *bufferedReader) ReadByte() byte {
	if br.err != nil {
		return 0
	}

	var b byte
	b, br.err = br.buf.ReadByte() // Read the byte and store any error.

	return b
}

// Next returns the next n bytes from the buffer without consuming them.
// If there are fewer than n bytes, it returns the entire buffer.
// The returned slice is only valid until the next read or write operation.
func (br *bufferedReader) Next(n int) []byte {
	if br.err != nil {
		return nil
	}

	var b []byte
	b, br.err = br.next(n) // Delegate to the internal next method.

	return b
}

// next is an internal helper method for Next.
// It peeks at the next n bytes and discards them after reading.
func (br *bufferedReader) next(n int) ([]byte, error) {
	if n == 0 {
		return nil, nil
	}

	// Peek at the next n bytes without consuming them.
	peeked, err := br.buf.Peek(n)
	if err != nil {
		return nil, err
	}

	// Discard the peeked bytes to advance the buffer.
	if _, err = br.buf.Discard(n); err != nil {
		return nil, err
	}

	return peeked, nil
}

// readTillDelimiter reads bytes from the buffer until the specified delimiter is found.
// It returns the data up to but not including the delimiter.
// If an error occurs before finding the delimiter, it returns the data read so far and the error.
func (br *bufferedReader) readTillDelimiter(delimiter byte) ([]byte, error) {
	read, err := br.buf.ReadBytes(delimiter)
	if err != nil || len(read) == 0 {
		return read, err
	}

	// Unread the delimiter so it's not consumed.
	err = br.buf.UnreadByte()

	return read[:len(read)-1], err
}

// readTillDelimiters reads bytes from the buffer until one of the specified delimiters is found.
// It returns the data up to but not including the delimiters.
// If an error occurs before finding the delimiters, it returns the data read so far and the error.
func (br *bufferedReader) readTillDelimiters(delimiters []byte) ([]byte, error) {
	if len(delimiters) == 0 {
		return nil, nil
	}

	// If there's only one delimiter, use the simpler readTillDelim method.
	if len(delimiters) == 1 {
		return br.readTillDelimiter(delimiters[0])
	}

	result := make([]byte, 0)

	for {
		// Read until the first delimiter.
		read, err := br.readTillDelimiter(delimiters[0])
		if err != nil {
			return result, err
		}

		result = append(result, read...)

		// Peek ahead to check if the full delimiter sequence matches.
		peeked, err := br.buf.Peek(len(delimiters))
		if err != nil {
			return result, err
		}

		// If the full delimiter sequence is found, stop reading.
		if bytes.Equal(peeked, delimiters) {
			break
		}

		// If not, read the next byte and continue.
		b, err := br.buf.ReadByte()
		if err != nil {
			return result, err
		}

		result = append(result, b)
	}

	return result, nil
}

// ReadText reads text from the buffer until the specified encoding's termination bytes are found.
// It discards the termination bytes and returns the text.
// This is useful for reading text fields in ID3v2 frames, which are often null-terminated.
func (br *bufferedReader) ReadText(encoding Encoding) []byte {
	if br.err != nil {
		return nil
	}

	var (
		text       []byte
		delimiters = encoding.TerminationBytes
	)

	// Read until the termination bytes are found.
	text, br.err = br.readTillDelimiters(delimiters)

	// Handle UTF-16 encoding edge case: if the text doesn't start with a BOM,
	// append the first byte to ensure proper decoding.
	if encoding.Equals(EncodingUTF16) &&
		!bytes.Equal(text, bom) {
		text = append(text, br.ReadByte())
	}

	// Discard the termination bytes.
	br.Discard(len(delimiters))

	return text
}

// Reset resets the bufferedReader to read from a new io.Reader.
// This is useful for reusing the bufferedReader with a different source.
func (br *bufferedReader) Reset(rd io.Reader) {
	br.buf.Reset(rd)
	br.err = nil
}
