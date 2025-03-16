package id3v2

import "io"

// UserDefinedTextFrame represents a TXXX frame in an ID3v2 tag.
// TXXX frames are used to store custom, user-defined text information.
// You can have multiple TXXX frames in a tag, but each one must have a unique description.
type UserDefinedTextFrame struct {
	Encoding    Encoding // The text encoding used for the description and value.
	Description string   // A unique description for this frame (e.g., "My Custom Field").
	Value       string   // The actual value associated with the description.
	Multi       []string // A slice of multiple values, if applicable (used for multi-value fields).
}

// Size calculates the total size of the UserDefinedTextFrame in bytes.
// This includes the encoding byte, the description, termination bytes, and the value.
func (udtf UserDefinedTextFrame) Size() int {
	return 1 + // 1 byte for the encoding
		encodedSize(udtf.Description, udtf.Encoding) + // Size of the description
		len(udtf.Encoding.TerminationBytes) + // Size of the termination bytes
		encodedSize(udtf.Value, udtf.Encoding) // Size of the value
}

// UniqueIdentifier returns a string that uniquely identifies this frame.
// For UserDefinedTextFrame, the description is used as the unique identifier
// since it must be unique within the tag.
func (udtf UserDefinedTextFrame) UniqueIdentifier() string {
	return udtf.Description
}

// WriteTo writes the UserDefinedTextFrame to the provided io.Writer.
// It returns the number of bytes written and any error encountered.
func (udtf UserDefinedTextFrame) WriteTo(w io.Writer) (n int64, err error) {
	return useBufferedWriter(w, func(bw *bufferedWriter) error {
		// Write the encoding byte.
		bw.WriteByte(udtf.Encoding.Key)

		// Write the description, encoded according to the specified encoding.
		bw.EncodeAndWriteText(udtf.Description, udtf.Encoding)

		// Write the termination bytes for the description.
		_, err = bw.Write(udtf.Encoding.TerminationBytes)
		if err != nil {
			return err
		}

		// Write the value, encoded according to the specified encoding.
		bw.EncodeAndWriteText(udtf.Value, udtf.Encoding)

		return nil
	})
}

// parseUserDefinedTextFrame parses a UserDefinedTextFrame from a bufferedReader.
// It reads the encoding, description, and value from the reader and constructs
// a UserDefinedTextFrame. If the frame contains multiple values, they are stored
// in the Multi field.
func parseUserDefinedTextFrame(br *bufferedReader, _ byte) (Framer, error) {
	// Read the encoding byte and determine the text encoding.
	encoding := getEncoding(br.ReadByte())

	// Read the description using the specified encoding.
	description := br.ReadText(encoding)

	// Check for errors after reading the description.
	if br.Err() != nil {
		return nil, br.Err()
	}

	// Use a buffer to read the value from the reader.
	value := getBytesBuffer()
	defer putBytesBuffer(value) // Ensure the buffer is returned to the pool.

	// Read the value from the reader into the buffer.
	if _, err := value.ReadFrom(br); err != nil {
		return nil, err
	}

	// Decode the value into a slice of strings, handling multi-value fields.
	values := decodeMulti(value.Bytes(), encoding)

	// Extract the first value if multiple values are present.
	var first string
	if len(values) > 0 {
		first = values[0]
	}

	// Construct and return the UserDefinedTextFrame.
	udtf := UserDefinedTextFrame{
		Encoding:    encoding,
		Description: decodeText(description, encoding),
		Value:       first,
		Multi:       values,
	}

	return udtf, nil
}
