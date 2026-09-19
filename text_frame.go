package id3v2

import "io"

func textValues(primary string, multi []string) []string {
	if len(multi) != 0 {
		return multi
	}

	return []string{primary}
}

// TextFrame is used to work with all text frames in ID3v2 tags.
// These frames are identified by IDs starting with "T" (e.g., TIT2 for title, TALB for album).
// It stores text data along with its encoding and supports multiple values for certain frames.
type TextFrame struct {
	Encoding Encoding // The encoding used for the text (e.g., UTF-8, ISO-8859-1).
	Text     string   // The primary text value of the frame.
	Multi    []string // Additional text values, used for frames that support multiple entries.
}

// textFrameUniqueIdentifier is a constant used to uniquely identify text frames.
// Since text frames don't have a unique identifier in the ID3v2 spec, this is a placeholder.
const textFrameUniqueIdentifier = "ID"

// Size calculates the total size of the TextFrame in bytes.
// This includes the encoding byte, the encoded text, and the termination bytes.
func (tf TextFrame) Size() int {
	return tf.sizeForVersion(4)
}

func (tf TextFrame) sizeForVersion(version byte) int {
	values := []string{tf.Text}
	if version == 4 {
		values = textValues(tf.Text, tf.Multi)
	}

	size := 1
	for _, value := range values {
		size += encodedSize(value, tf.Encoding) + len(tf.Encoding.TerminationBytes)
	}

	return size
}

// UniqueIdentifier returns a unique identifier for the TextFrame.
// Since text frames don't have a unique identifier in the ID3v2 spec, this returns a constant.
func (tf TextFrame) UniqueIdentifier() string {
	return textFrameUniqueIdentifier
}

// WriteTo writes the TextFrame to the provided io.Writer.
// It encodes the text using the specified encoding and writes the frame's data.
// Returns the number of bytes written and any error encountered.
func (tf TextFrame) WriteTo(w io.Writer) (int64, error) {
	return tf.writeToVersion(w, 4)
}

func (tf TextFrame) writeToVersion(w io.Writer, version byte) (int64, error) {
	return useBufferedWriter(w, func(bw *bufferedWriter) error {
		// Write the encoding byte.
		bw.writeByte(tf.Encoding.Key)

		values := []string{tf.Text}
		if version == 4 {
			values = textValues(tf.Text, tf.Multi)
		}

		for _, value := range values {
			// Encode and write the text using the specified encoding.
			bw.EncodeAndWriteText(value, tf.Encoding)

			// Write the termination bytes for the encoding.
			if _, err := bw.Write(tf.Encoding.TerminationBytes); err != nil {
				return err
			}
		}

		return nil
	})
}

// parseTextFrame parses a TextFrame from a bufferedReader.
// It reads the encoding, text, and any additional values from the reader.
// Returns a TextFrame and any error encountered during parsing.
func parseTextFrame(br *bufferedReader) (Framer, error) {
	// Read the encoding byte and determine the encoding type.
	encoding := getEncoding(br.readByte())

	// Check for errors after reading the encoding byte.
	if br.Err() != nil {
		return nil, br.Err()
	}

	// Get a buffer to store the raw text data.
	buf := getBytesBuffer()
	defer putBytesBuffer(buf) // Ensure the buffer is returned to the pool after use.

	// Read the rest of the frame's data into the buffer.
	if _, err := buf.ReadFrom(br); err != nil {
		return nil, err
	}

	// Decode the raw data into a slice of strings, handling multi-value frames.
	values := decodeMulti(buf.Bytes(), encoding)

	// Extract the first value as the primary text.
	var first string
	if len(values) > 0 {
		first = values[0]
	}

	// Create and return the TextFrame.
	tf := TextFrame{
		Encoding: encoding,
		Text:     first,
		Multi:    values,
	}

	return tf, nil
}
