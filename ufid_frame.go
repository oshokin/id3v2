package id3v2

import (
	"errors"
	"fmt"
	"io"
)

const ufidMaxIdentifierSize = 64

// UFIDFrame represents a "Unique File Identifier" frame in an ID3v2 tag.
// This frame is used to store a unique identifier for the file, typically associated with
// an owner (e.g., a service like MusicBrainz) and a binary identifier.
type UFIDFrame struct {
	OwnerIdentifier string // The owner of the unique identifier (e.g., "https://musicbrainz.org").
	Identifier      []byte // The unique identifier itself, stored as a byte slice.
}

// UniqueIdentifier returns a string that uniquely identifies this frame.
// For UFID frames, this is the OwnerIdentifier, as it distinguishes this frame from others.
func (ufid UFIDFrame) UniqueIdentifier() string {
	return ufid.OwnerIdentifier
}

func (ufid UFIDFrame) encodedOwner() int {
	return encodedSize(ufid.OwnerIdentifier, EncodingISO) + len(EncodingISO.TerminationBytes)
}

// Size calculates the total size of the UFID frame in bytes.
// This includes the size of the owner identifier (encoded in ISO-8859-1), the termination bytes,
// and the size of the identifier itself.
func (ufid UFIDFrame) Size() int {
	return ufid.encodedOwner() + len(ufid.Identifier)
}

// WriteTo writes the UFID frame to the provided io.Writer.
// Size and WriteTo use the same ISO-8859-1 encoding path for the owner identifier.
// The frame is written in the following format:
//   - Owner identifier (encoded in ISO-8859-1)
//   - Termination bytes (0x00 for ISO-8859-1)
//   - Identifier (raw bytes)
func (ufid UFIDFrame) WriteTo(w io.Writer) (n int64, err error) {
	if ufid.OwnerIdentifier == "" {
		return 0, errors.New("UFID owner identifier must not be empty")
	}

	if len(ufid.Identifier) > ufidMaxIdentifierSize {
		return 0, fmt.Errorf(
			"UFID identifier is %d bytes; maximum is %d",
			len(ufid.Identifier),
			ufidMaxIdentifierSize,
		)
	}

	return useBufferedWriter(w, func(bw *bufferedWriter) error {
		// Write the owner identifier encoded as ISO-8859-1.
		bw.EncodeAndWriteText(ufid.OwnerIdentifier, EncodingISO)

		// Write the termination bytes for the owner identifier.
		if _, err := bw.Write(EncodingISO.TerminationBytes); err != nil {
			return err
		}

		// Write the identifier as raw bytes.
		if _, err := bw.Write(ufid.Identifier); err != nil {
			return err
		}

		return nil
	})
}

// parseUFIDFrame parses a UFID frame from a bufferedReader.
// It reads the owner identifier and the unique identifier from the reader and constructs a UFIDFrame.
// The `_ byte` parameter is unused but required to match the Framer interface.
func parseUFIDFrame(br *bufferedReader, _ byte) (Framer, error) {
	// Read the owner identifier, which is encoded in ISO-8859-1.
	owner := br.ReadText(EncodingISO)

	// Read the remaining bytes as the unique identifier.
	ident := br.ReadAll()

	if br.Err() != nil {
		return nil, br.Err()
	}

	return UFIDFrame{
		OwnerIdentifier: decodeText(owner, EncodingISO), // Decode the owner identifier from ISO-8859-1 to a string.
		Identifier:      ident,                          // Use the raw bytes for the identifier.
	}, nil
}
