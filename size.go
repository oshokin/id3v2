package id3v2

import "errors"

const (
	// id3SizeLen is the length of the ID3v2 size format, which is 4 bytes (4 * 0bxxxxxxxx).
	id3SizeLen = 4

	// synchSafeMaxSize is the maximum allowed size for a synch-safe integer in ID3v2 tags.
	// Synch-safe integers are used to avoid false synchronization in MP3 streams.
	synchSafeMaxSize = 268435455 // == 0b00001111 11111111 11111111 11111111

	// synchSafeSizeBase is the number of bits used per byte in a synch-safe integer.
	synchSafeSizeBase = 7 // == 0b01111111

	// synchSafeMask is a bitmask used to extract the first 7 bits of a 32-bit integer.
	synchSafeMask = uint(254 << (3 * 8)) // 11111110 000000000 000000000 000000000

	// synchUnsafeMaxSize is the maximum allowed size for a non-synch-safe integer in ID3v2 tags.
	synchUnsafeMaxSize = 4294967295 // == 0b11111111 11111111 11111111 11111111

	// synchUnsafeSizeBase is the number of bits used per byte in a non-synch-safe integer.
	synchUnsafeSizeBase = 8 // == 0b11111111

	// synchUnsafeMask is a bitmask used to extract the first 8 bits of a 32-bit integer.
	synchUnsafeMask = uint(255 << (3 * 8)) // 11111111 000000000 000000000 000000000
)

var (
	// ErrInvalidSizeFormat is returned when the size format of a tag or frame is invalid.
	ErrInvalidSizeFormat = errors.New("invalid format of tag's/frame's size")

	// ErrSizeOverflow is returned when the size of a tag or frame exceeds the maximum allowed size.
	ErrSizeOverflow = errors.New("size of tag/frame is greater than allowed in id3 tag")
)

// writeBytesSize writes the size of a tag or frame to a bufferedWriter.
// It handles both synch-safe and non-synch-safe sizes.
func writeBytesSize(bw *bufferedWriter, size uint, synchSafe bool) error {
	if synchSafe {
		return writeSynchSafeBytesSize(bw, size)
	}

	return writeSynchUnsafeBytesSize(bw, size)
}

// writeSynchSafeBytesSize writes a synch-safe size to a bufferedWriter.
// Synch-safe sizes are used to avoid false synchronization in MP3 streams.
func writeSynchSafeBytesSize(bw *bufferedWriter, size uint) error {
	// Check if the size exceeds the maximum allowed for synch-safe integers.
	if size > synchSafeMaxSize {
		return ErrSizeOverflow
	}

	// Shift the size left by 4 bits to skip the first 4 bits, which are always "0"
	// in synch-safe integers. This ensures the size fits within the allowed range.
	size <<= 4

	// The algorithm works by processing the size in chunks of 7 bits per byte.
	// For example, if the size is a 32-bit integer like "10100111 01110101 01010010 11110000",
	// after skipping the first 4 bits, it becomes "10100111 01110101 01010010 11110000".
	// We then extract and write the first 7 bits of this value in each iteration.
	for range id3SizeLen {
		// Extract the first 7 bits of the size using a bitmask.
		firstBits := size & synchSafeMask
		// Shift the extracted bits to the least significant byte position.
		// This is necessary because we need to convert the 7 bits into a single byte.
		firstBits >>= (3*8 + 1)
		// Convert the shifted bits to a byte.
		bSize := byte(firstBits)
		// Write the byte to the bufferedWriter.
		bw.WriteByte(bSize)
		// Shift the size left by 7 bits to process the next 7 bits in the next iteration.
		size <<= synchSafeSizeBase
	}

	return nil
}

// writeSynchUnsafeBytesSize writes a non-synch-safe size to a bufferedWriter.
// Non-synch-safe sizes are used when synchronization is not a concern.
func writeSynchUnsafeBytesSize(bw *bufferedWriter, size uint) error {
	if size > synchUnsafeMaxSize {
		return ErrSizeOverflow
	}

	// Write the size in 4 bytes, each containing 8 bits of the size.
	for range id3SizeLen {
		// Extract the first 8 bits of the size.
		firstBits := size & synchUnsafeMask
		// Shift the extracted bits to the least significant byte position.
		firstBits >>= (3 * 8)
		// Convert the bits to a byte and write it to the bufferedWriter.
		bw.WriteByte(byte(firstBits))
		// Shift the size left by 8 bits to process the next 8 bits.
		size <<= synchUnsafeSizeBase
	}

	return nil
}

// parseSize parses the size of a tag or frame from a byte slice.
// It handles both synch-safe and non-synch-safe sizes.
func parseSize(data []byte, synchSafe bool) (int64, error) {
	if len(data) > id3SizeLen {
		return 0, ErrInvalidSizeFormat
	}

	// Determine the number of bits per byte based on whether the size is synch-safe.
	var sizeBase uint
	if synchSafe {
		sizeBase = synchSafeSizeBase
	} else {
		sizeBase = synchUnsafeSizeBase
	}

	var size int64

	// Parse each byte of the size.
	for _, b := range data {
		// For synch-safe sizes, ensure that the most significant bit is not set.
		if synchSafe && b&128 > 0 { // 128 = 0b1000_0000
			return 0, ErrInvalidSizeFormat
		}

		// Shift the current size left by the number of bits per byte and add the new byte.
		size = (size << sizeBase) | int64(b)
	}

	return size, nil
}
