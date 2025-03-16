package id3v2

import (
	"errors"
	"io"
)

// ErrInvalidLanguageLength is returned when a language code does not meet the ISO 639-2 standard,
// which requires language codes to be exactly three letters long.
var ErrInvalidLanguageLength = errors.New("language code must consist of three letters according to ISO 639-2")

// Framer is an interface that defines the behavior of an ID3v2 frame.
// Any custom frame implementation must satisfy this interface to be compatible with the ID3v2 package.
type Framer interface {
	// Size returns the size of the frame's body in bytes.
	Size() int

	// UniqueIdentifier returns a string that uniquely distinguishes this frame from others
	// with the same frame ID. For example, frames like TXXX or APIC can have multiple instances
	// in a tag, differentiated by properties like descriptions or picture types.
	//
	// For frames that can only appear once with the same ID (e.g., text frames), this method
	// should return an empty string ("").
	UniqueIdentifier() string

	// WriteTo writes the frame's body to the provided io.Writer.
	// It returns the number of bytes written and any error encountered during the write operation.
	WriteTo(w io.Writer) (n int64, err error)
}
