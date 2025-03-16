package id3v2

import (
	"io"
	"math/big"
)

// PopularimeterFrame represents a Popularimeter (POPM) frame in an ID3v2 tag.
// The Popularimeter frame is used to store user-specific play count and rating information for a track.
// For more details, see: https://id3.org/id3v2.3.0#Popularimeter
type PopularimeterFrame struct {
	// Email is the identifier for the POPM frame. It typically represents the email address
	// of the user who rated or played the track.
	Email string

	// Rating is the user's rating for the track. It ranges from 1 to 255, where:
	// - 1 is the worst rating.
	// - 255 is the best rating.
	// - 0 means the rating is unknown.
	Rating uint8

	// Counter is the number of times the track has been played by the user identified by the Email field.
	// It is stored as a big.Int to support very large play counts.
	Counter *big.Int
}

// UniqueIdentifier returns a unique identifier for the PopularimeterFrame.
// For POPM frames, the unique identifier is the Email field, as each email address
// corresponds to a unique user's rating and play count.
func (pf PopularimeterFrame) UniqueIdentifier() string {
	return pf.Email
}

// Size calculates the total size of the PopularimeterFrame in bytes.
// This includes the size of the Email, Rating, and Counter fields.
func (pf PopularimeterFrame) Size() int {
	ratingSize := 1 // The Rating field is always 1 byte.

	// Total size = Email length + 1 (null terminator) + Rating size + Counter size.
	return len(pf.Email) + 1 + ratingSize + len(pf.counterBytes())
}

// counterBytes converts the Counter field into a byte slice.
// The ID3v2 specification requires the counter to be at least 4 bytes long.
// If the counter is smaller than 4 bytes, it is padded with leading zeros.
func (pf PopularimeterFrame) counterBytes() []byte {
	bytes := pf.Counter.Bytes() // Get the byte representation of the counter.

	// If the counter is less than 4 bytes, pad it with leading zeros.
	bytesNeeded := 4 - len(bytes)
	if bytesNeeded > 0 {
		padding := make([]byte, bytesNeeded)
		bytes = append(padding, bytes...)
	}

	return bytes
}

// WriteTo writes the PopularimeterFrame to the provided io.Writer.
// It returns the number of bytes written and any error encountered during the write operation.
func (pf PopularimeterFrame) WriteTo(w io.Writer) (n int64, err error) {
	return useBufferedWriter(w, func(bw *bufferedWriter) error {
		// Write the Email field, followed by a null terminator (0).
		bw.WriteString(pf.Email)
		bw.WriteByte(0)

		// Write the Rating field.
		bw.WriteByte(pf.Rating)

		// Write the Counter field as a 4-byte value.
		_, err = bw.Write(pf.counterBytes())
		if err != nil {
			return err
		}

		return nil
	})
}

// parsePopularimeterFrame parses a PopularimeterFrame from a bufferedReader.
// It reads the Email, Rating, and Counter fields from the reader and constructs a PopularimeterFrame.
//
//nolint:unparam // Error is intentionally nil to satisfy the framers map function contract.
func parsePopularimeterFrame(br *bufferedReader, _ byte) (Framer, error) {
	// Read the Email field as ISO-8859-1 encoded text.
	email := br.ReadText(EncodingISO)

	// Read the Rating field as a single byte.
	rating := br.ReadByte()

	// Read the remaining bytes as the Counter field.
	remainingBytes := br.ReadAll()

	// Convert the remaining bytes into a big.Int to represent the play count.
	counter := big.NewInt(0)
	counter = counter.SetBytes(remainingBytes)

	// Construct and return the PopularimeterFrame.
	pf := PopularimeterFrame{
		Email:   string(email),
		Rating:  rating,
		Counter: counter,
	}

	return pf, nil
}
