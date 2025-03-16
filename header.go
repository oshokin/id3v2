package id3v2

import (
	"bytes"
	"errors"
	"io"
)

// tagHeaderSize is the size of an ID3v2 tag header in bytes.
const tagHeaderSize = 10

var (
	// ErrSmallHeaderSize is returned when the size of the tag header is smaller than expected.
	ErrSmallHeaderSize = errors.New("size of tag header is less than expected")

	// id3Identifier is the magic number that identifies an ID3v2 tag.
	id3Identifier = []byte("ID3")

	// ErrNoTag is returned when the file does not contain an ID3v2 tag.
	ErrNoTag = errors.New("there is no tag in file")
)

// tagHeader represents the header of an ID3v2 tag.
// It contains the size of the frames and the version of the ID3v2 tag.
type tagHeader struct {
	FramesSize int64 // Size of the frames in bytes.
	Version    byte  // Version of the ID3v2 tag (e.g., 3 for ID3v2.3, 4 for ID3v2.4).
}

// parseHeader reads and parses the ID3v2 tag header from the provided reader.
// It returns a tagHeader struct containing the parsed information.
// If the reader does not contain an ID3v2 tag, it returns errNoTag.
// If the reader provides fewer bytes than the expected header size, it returns ErrSmallHeaderSize.
func parseHeader(rd io.Reader) (tagHeader, error) {
	var header tagHeader

	// Create a buffer to hold the tag header data.
	data := make([]byte, tagHeaderSize)

	// Read the tag header from the reader.
	n, err := rd.Read(data)
	if err != nil {
		return header, err
	}

	// Check if the number of bytes read is less than the expected header size.
	if n < tagHeaderSize {
		return header, ErrSmallHeaderSize
	}

	// Check if the data starts with the ID3 identifier.
	if !isID3Tag(data[0:3]) {
		return header, ErrNoTag
	}

	// Extract the version of the ID3v2 tag from the header.
	header.Version = data[3]

	// Parse the size of the frames from the header.
	// The size is stored in a synchsafe format, which ensures that the most significant bit of each byte is 0.
	size, err := parseSize(data[6:], true)
	if err != nil {
		return header, err
	}

	// Store the parsed size in the header struct.
	header.FramesSize = size

	return header, nil
}

// isID3Tag checks if the provided data starts with the ID3 identifier.
// It returns true if the data matches the ID3 identifier, otherwise false.
func isID3Tag(data []byte) bool {
	return len(data) == len(id3Identifier) && bytes.Equal(data, id3Identifier)
}
