package id3v2

import (
	"bufio"
	"io"
	"math"
)

// truncateIntToUint converts an int to a uint, ensuring it doesn't overflow.
// If the input value is negative, it returns 0 to prevent invalid uint values.
func truncateIntToUint(value int) uint {
	if value < 0 {
		return 0
	}

	return uint(value)
}

// truncateInt64ToInt32 safely truncates a 64-bit integer to a 32-bit integer.
// If the value exceeds the range of int32, it is clamped to the nearest valid value.
func truncateInt64ToInt32(value int64) int32 {
	if value < math.MinInt32 {
		return math.MinInt32
	}

	if value > math.MaxInt32 {
		return math.MaxInt32
	}

	return int32(value)
}

// truncateInt64ToUint32 safely truncates a 64-bit integer to a 32-bit unsigned integer.
// If the value exceeds the range of uint32, it returns math.MaxUint32.
// Otherwise, it returns the value as a uint32.
func truncateInt64ToUint32(value int64) uint32 {
	if value < 0 {
		return 0
	}

	if value > math.MaxUint32 {
		return math.MaxUint32
	}

	return uint32(value) //nolint:gosec // The value is already validated above.
}

// truncateUintToInt64 safely truncates a uint to a 64-bit signed integer.
// If the value exceeds the range of int64, it returns math.MaxInt64.
// Otherwise, it returns the value as an int64.
func truncateUintToInt64(value uint) int64 {
	if value > math.MaxInt64 {
		return math.MaxInt64
	}

	return int64(value)
}

// readLinesFromReader reads lines from an io.Reader and applies a transformation function to each line.
// The transform function takes a line as input and returns the transformed line and a boolean indicating
// whether the line should be skipped (true to skip, false to include).
// Returns a slice of transformed lines or an error if reading fails.
func readLinesFromReader(inputReader io.Reader, transform func(string) (string, bool)) ([]string, error) {
	var (
		lines   []string
		scanner = bufio.NewScanner(inputReader)
	)

	for scanner.Scan() {
		line, isLineSkipped := transform(scanner.Text())
		if isLineSkipped {
			continue
		}

		lines = append(lines, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}
