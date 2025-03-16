package id3v2

import "io"

// LinkFrame represents a frame that contains a URL or link.
// It is used for frames like "WXXX" (User-defined URL link) in ID3v2 tags.
type LinkFrame struct {
	Encoding Encoding // The text encoding used for the URL.
	URL      string   // The actual URL or link.
}

// linkFrameUniqueIdentifier is a constant used to uniquely identify LinkFrame instances.
// Since LinkFrame doesn't have a natural unique identifier, this constant is used.
const linkFrameUniqueIdentifier = "ID"

// Size calculates the total size of the LinkFrame in bytes.
// This includes the encoding byte, the encoded URL, and the termination bytes.
func (lf LinkFrame) Size() int {
	return 1 + encodedSize(lf.URL, lf.Encoding) + len(lf.Encoding.TerminationBytes)
}

// UniqueIdentifier returns a unique identifier for the LinkFrame.
// Since LinkFrame doesn't have a natural unique identifier, it uses a constant value.
func (lf LinkFrame) UniqueIdentifier() string {
	return linkFrameUniqueIdentifier
}

// WriteTo writes the LinkFrame to the provided io.Writer.
// It encodes the URL using the specified encoding and writes the frame's data.
// Returns the number of bytes written and any error encountered.
func (lf LinkFrame) WriteTo(w io.Writer) (int64, error) {
	return useBufferedWriter(w, func(bw *bufferedWriter) error {
		// Write the encoding byte.
		bw.WriteByte(lf.Encoding.Key)

		// Encode and write the URL.
		bw.EncodeAndWriteText(lf.URL, lf.Encoding)

		// Write the termination bytes for the encoding.
		_, err := bw.Write(lf.Encoding.TerminationBytes)
		if err != nil {
			return err
		}

		return nil
	})
}

// parseLinkFrame parses a LinkFrame from a bufferedReader.
// It reads the encoding, URL, and termination bytes, and constructs a LinkFrame.
// Returns the parsed LinkFrame and any error encountered.
func parseLinkFrame(br *bufferedReader) (Framer, error) {
	// Read the encoding byte and determine the encoding type.
	encoding := getEncoding(br.ReadByte())

	// Check for errors after reading the encoding byte.
	if br.Err() != nil {
		return nil, br.Err()
	}

	// Get a reusable bytes.Buffer from the pool to store the URL.
	buf := getBytesBuffer()
	defer putBytesBuffer(buf) // Return the buffer to the pool when done.

	// Read the remaining data (URL) from the bufferedReader into the buffer.
	if _, err := buf.ReadFrom(br); err != nil {
		return nil, err
	}

	// Decode the URL from the buffer using the specified encoding.
	lf := LinkFrame{
		Encoding: encoding,
		URL:      decodeText(buf.Bytes(), encoding),
	}

	return lf, nil
}
