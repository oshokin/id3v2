package id3v2

import (
	"bytes"
	"io"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/unicode"
)

// Encoding represents a text encoding used in ID3v2 tags.
// It includes the encoding name, a key (used in ID3v2 frames), and the termination bytes
// that mark the end of a string in this encoding.
type Encoding struct {
	Name             string // Name of the encoding (e.g., "ISO-8859-1").
	Key              byte   // Key used in ID3v2 frames to identify this encoding.
	TerminationBytes []byte // Bytes that mark the end of a string in this encoding.
}

// Equals checks if this Encoding is equal to another Encoding by comparing their keys.
func (e Encoding) Equals(other Encoding) bool {
	return e.Key == other.Key
}

// String returns the name of the encoding.
func (e Encoding) String() string {
	return e.Name
}

// Available encodings supported by ID3v2.
var (
	// EncodingISO is ISO-8859-1 encoding, commonly used in older ID3v2 tags.
	EncodingISO = Encoding{
		Name:             "ISO-8859-1",
		Key:              0,
		TerminationBytes: []byte{0},
	}

	// EncodingUTF16 is UTF-16 encoded Unicode with a Byte Order Mark (BOM).
	EncodingUTF16 = Encoding{
		Name:             "UTF-16 encoded Unicode with BOM",
		Key:              1,
		TerminationBytes: []byte{0, 0},
	}

	// EncodingUTF16BE is UTF-16BE encoded Unicode without a BOM.
	EncodingUTF16BE = Encoding{
		Name:             "UTF-16BE encoded Unicode without BOM",
		Key:              2,
		TerminationBytes: []byte{0, 0},
	}

	// EncodingUTF8 is UTF-8 encoded Unicode, the most widely used encoding.
	EncodingUTF8 = Encoding{
		Name:             "UTF-8 encoded Unicode",
		Key:              3,
		TerminationBytes: []byte{0},
	}

	// encodings is a slice of all supported encodings.
	encodings = []Encoding{EncodingISO, EncodingUTF16, EncodingUTF16BE, EncodingUTF8}

	// xEncodingISO is the Go encoding for ISO-8859-1.
	xEncodingISO = charmap.ISO8859_1

	// xEncodingUTF16BEBOM is the Go encoding for UTF-16 with Big Endian and BOM.
	xEncodingUTF16BEBOM = unicode.UTF16(unicode.BigEndian, unicode.ExpectBOM)

	// xEncodingUTF16LEBOM is the Go encoding for UTF-16 with Little Endian and BOM.
	xEncodingUTF16LEBOM = unicode.UTF16(unicode.LittleEndian, unicode.ExpectBOM)

	// xEncodingUTF16BE is the Go encoding for UTF-16 with Big Endian and no BOM.
	xEncodingUTF16BE = unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM)

	// xEncodingUTF8 is the Go encoding for UTF-8.
	xEncodingUTF8 = unicode.UTF8
)

var (
	// bom is the little-endian Byte Order Mark used in UTF-16 encoded Unicode.
	// See https://en.wikipedia.org/wiki/Byte_order_mark.
	bom = []byte{0xFF, 0xFE}

	// utf16BEBOM is the big-endian Byte Order Mark used in UTF-16 encoded Unicode.
	utf16BEBOM = []byte{0xFE, 0xFF}
)

// getEncoding returns the Encoding corresponding to the given ID3v2 key.
// If the key is invalid, it defaults to EncodingUTF8.
func getEncoding(key byte) Encoding {
	if key > 3 {
		return EncodingUTF8
	}

	return encodings[key]
}

// encodedSize calculates the length of the UTF-8 string `src` when encoded into the specified `enc`.
// If the encoding is already UTF-8, it returns the length of the string as is.
func encodedSize(src string, enc Encoding) int {
	if enc.Equals(EncodingUTF8) {
		return len(src)
	}

	// Use a buffered writer to calculate the encoded size.
	bw := getBufWriter(io.Discard)
	defer putBufWriter(bw)

	err := encodeWriteText(bw, src, enc)
	if err != nil {
		panic(err) // Panic if encoding fails, as this should never happen in normal usage.
	}

	return bw.Written()
}

// decodeText decodes the byte slice `src` from the specified `from` encoding into a UTF-8 string.
// It removes the termination bytes and handles special cases like BOM in UTF-16.
func decodeText(src []byte, from Encoding) string {
	src = bytes.TrimSuffix(src, from.TerminationBytes) // Remove termination bytes.

	if from.Equals(EncodingUTF8) {
		return string(src) // No decoding needed for UTF-8.
	}

	// If the source is just a BOM, return an empty string.
	if from.Equals(EncodingUTF16) && bytes.Equal(src, bom) {
		return ""
	}

	// Resolve the Go encoding for the specified ID3v2 encoding.
	fromXEncoding := resolveXEncoding(src, from)

	// Decode the byte slice into a UTF-8 string.
	// Invalid UTF-16 sequences become U+FFFD via x/text; a real U+FFFD in the
	// source text must be kept, so replacement characters are never stripped.
	result, err := fromXEncoding.NewDecoder().Bytes(src)
	if err != nil {
		return string(src) // Fallback to raw bytes if decoding fails.
	}

	return string(result)
}

// decodeMulti decodes a multi-valued byte slice `src` from the specified `from` encoding into a slice of UTF-8 strings.
// ISO-8859-1 and UTF-8 values are split on a single 0x00.
// UTF-16 and UTF-16BE values are split on a 0x00 0x00 terminator only at a two-byte code-unit boundary.
func decodeMulti(src []byte, from Encoding) []string {
	parts := splitEncodedTextValues(src, from)
	res := make([]string, 0, len(parts))

	if from.Equals(EncodingUTF16) {
		bomPrefix := utf16BOMPrefix(parts)
		for i, part := range parts {
			if i > 0 && !hasUTF16BOM(part) {
				part = append(append(make([]byte, 0, len(bomPrefix)+len(part)), bomPrefix...), part...)
			}

			res = append(res, decodeText(part, from))
		}

		return res
	}

	for _, part := range parts {
		res = append(res, decodeText(part, from))
	}

	return res
}

// splitEncodedTextValues splits encoded text values on the encoding terminator.
func splitEncodedTextValues(src []byte, from Encoding) [][]byte {
	if from.Equals(EncodingUTF16) || from.Equals(EncodingUTF16BE) {
		return splitUTF16TextValues(src)
	}

	src = bytes.TrimSuffix(src, from.TerminationBytes)

	return bytes.Split(src, from.TerminationBytes)
}

// splitUTF16TextValues splits UTF-16 text on a 0x00 0x00 terminator only when it is aligned to a two-byte code unit.
func splitUTF16TextValues(src []byte) [][]byte {
	if n := len(src); n >= 2 && n%2 == 0 && src[n-2] == 0 && src[n-1] == 0 {
		src = src[:n-2]
	}

	parts := make([][]byte, 0, 2)
	start := 0

	for i := 0; i+1 < len(src); i += 2 {
		if src[i] == 0 && src[i+1] == 0 {
			parts = append(parts, src[start:i])
			start = i + 2
		}
	}

	return append(parts, src[start:])
}

// hasUTF16BOM reports whether src begins with a UTF-16 BOM.
func hasUTF16BOM(src []byte) bool {
	return len(src) >= 2 && (bytes.Equal(src[:2], bom) || bytes.Equal(src[:2], utf16BEBOM))
}

// utf16BOMPrefix returns the BOM of the first UTF-16 value, or the big-endian BOM used when writing EncodingUTF16.
func utf16BOMPrefix(parts [][]byte) []byte {
	if len(parts) > 0 && hasUTF16BOM(parts[0]) {
		return parts[0][:2]
	}

	return utf16BEBOM
}

// encodeWriteText encodes the UTF-8 string `src`
// into the specified `to` encoding and writes it to the buffered writer `bw`.
// It handles special cases like adding a null terminator for UTF-16.
func encodeWriteText(bw *bufferedWriter, src string, to Encoding) error {
	if to.Equals(EncodingUTF8) {
		bw.WriteString(src) // No encoding needed for UTF-8.

		return nil
	}

	// Resolve the Go encoding for the specified ID3v2 encoding.
	toXEncoding := resolveXEncoding(nil, to)

	// Encode the string into the target encoding.
	encoded, err := toXEncoding.NewEncoder().String(src)
	if err != nil {
		return err
	}

	bw.WriteString(encoded)

	return nil
}

// resolveXEncoding resolves the Go encoding for the specified ID3v2 encoding.
// It handles special cases like detecting the BOM in UTF-16.
func resolveXEncoding(src []byte, encoding Encoding) encoding.Encoding {
	switch encoding.Key {
	case 0:
		return xEncodingISO // ISO-8859-1.
	case 1:
		// If the source starts with a BOM, use Little Endian; otherwise, use Big Endian.
		if len(src) > 2 && bytes.Equal(src[:2], bom) {
			return xEncodingUTF16LEBOM
		}

		return xEncodingUTF16BEBOM
	case 2:
		return xEncodingUTF16BE // UTF-16 Big Endian without BOM.
	default:
		return xEncodingUTF8 // Default to UTF-8.
	}
}
