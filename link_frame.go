package id3v2

import "io"

// LinkFrame represents a WXXX (user-defined URL link) frame.
// The public v2 model historically stores only Encoding and URL.
// The WXXX description field is parsed and discarded because adding
// an exported Description field would break unkeyed composite literals.
type LinkFrame struct {
	Encoding Encoding // The text encoding used for the description terminator.
	URL      string   // The actual URL or link.
}

// linkFrameUniqueIdentifier is a constant used to uniquely identify LinkFrame instances.
// Since LinkFrame doesn't have a natural unique identifier, this constant is used.
const linkFrameUniqueIdentifier = "ID"

// Size calculates the total size of the LinkFrame in bytes.
// This includes the encoding byte, an empty description terminator, and the ISO-8859-1 URL.
func (lf LinkFrame) Size() int {
	return 1 +
		len(lf.Encoding.TerminationBytes) +
		encodedSize(lf.URL, EncodingISO)
}

// UniqueIdentifier returns a unique identifier for the LinkFrame.
// Since LinkFrame doesn't have a natural unique identifier, it uses a constant value.
func (lf LinkFrame) UniqueIdentifier() string {
	return linkFrameUniqueIdentifier
}

// WriteTo writes the LinkFrame to the provided io.Writer.
// WXXX is serialized as: encoding, empty description terminator, URL in ISO-8859-1.
func (lf LinkFrame) WriteTo(w io.Writer) (int64, error) {
	return useBufferedWriter(w, func(bw *bufferedWriter) error {
		// Write the encoding byte.
		bw.writeByte(lf.Encoding.Key)

		// Write the empty description terminator required by WXXX.
		if _, err := bw.Write(lf.Encoding.TerminationBytes); err != nil {
			return err
		}

		// Encode and write the URL as ISO-8859-1.
		bw.EncodeAndWriteText(lf.URL, EncodingISO)

		return nil
	})
}

func parseUserDefinedURLFrame(br *bufferedReader, _ byte) (Framer, error) {
	return parseLinkFrame(br)
}

func parseLinkFrame(br *bufferedReader) (Framer, error) {
	encoding := getEncoding(br.readByte())
	if br.Err() != nil {
		return nil, br.Err()
	}

	// Description is required by the WXXX layout but is not represented
	// in the public v2 LinkFrame. Discard it to keep source compatibility.
	_ = br.ReadText(encoding)
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

	// Decode the URL from the buffer using ISO-8859-1.
	return LinkFrame{
		Encoding: encoding,
		URL:      decodeText(buf.Bytes(), EncodingISO),
	}, nil
}
