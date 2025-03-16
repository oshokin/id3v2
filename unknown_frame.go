package id3v2

import (
	"io"
	"math/rand/v2"
	"strconv"
)

// UnknownFrame represents an ID3v2 frame that the library doesn't know how to parse or interpret.
// It stores the raw byte data of the frame, allowing the library to handle unknown frame types
// without losing their content. This is useful for preserving custom or proprietary frames.
type UnknownFrame struct {
	Body []byte // Raw byte data of the unknown frame.
}

// UniqueIdentifier generates a unique identifier for the UnknownFrame.
// Since the frame type is unknown, this method uses a random integer to ensure uniqueness.
// This is necessary because ID3v2 frames typically have unique identifiers, but unknown frames
// don't have a predefined ID.
func (uf UnknownFrame) UniqueIdentifier() string {
	// Generate a random integer and convert it to a string to ensure uniqueness.
	return strconv.Itoa(rand.Int())
}

// Size returns the size of the UnknownFrame's body in bytes.
// This is used to calculate the total size of the frame when writing it to a file.
func (uf UnknownFrame) Size() int {
	return len(uf.Body)
}

// WriteTo writes the raw byte data of the UnknownFrame to the provided io.Writer.
// It returns the number of bytes written and any error encountered during the write operation.
// This method is used to serialize the UnknownFrame back into an ID3v2 tag.
func (uf UnknownFrame) WriteTo(w io.Writer) (n int64, err error) {
	i, err := w.Write(uf.Body) // Write the raw body to the writer.

	return int64(i), err // Return the number of bytes written and any error.
}

// parseUnknownFrame parses an unknown frame from a bufferedReader.
// It reads all remaining bytes from the reader and stores them in an UnknownFrame.
// This function is used when the library encounters a frame type it doesn't recognize.
func parseUnknownFrame(br *bufferedReader) (Framer, error) {
	body := br.ReadAll() // Read all remaining bytes from the bufferedReader.

	// Return an UnknownFrame containing the raw byte data and any error from the reader.
	return UnknownFrame{Body: body}, br.Err()
}
