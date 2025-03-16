package id3v2

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"code.cloudfoundry.org/bytefmt"
)

const (
	frameHeaderSize   = 10                    // Size of an ID3v2 frame header in bytes.
	defaultBufferSize = 32 * bytefmt.KILOBYTE // Default size of a byte buffer.
)

var (
	// ErrUnsupportedVersion is returned when the ID3v2 tag version is less than 3.
	ErrUnsupportedVersion = errors.New("unsupported version of ID3 tag")

	// ErrBodyOverflow is returned when a frame's size exceeds the remaining space in the tag.
	ErrBodyOverflow = errors.New("frame went over tag area")

	// ErrBlankFrame is returned when a frame's ID or size is empty or invalid.
	ErrBlankFrame = errors.New("id or size of frame are blank")
)

// frameHeader represents the header of an ID3v2 frame, containing the frame ID and body size.
type frameHeader struct {
	ID       string // The 4-character frame ID (e.g., "TIT2" for title).
	BodySize int64  // The size of the frame's body in bytes.
}

// parse reads the ID3v2 tag from the provided reader and parses it according to the given options.
// If the reader is smaller than expected, it returns ErrSmallHeaderSize.
func (tag *Tag) parse(rd io.Reader, opts Options) error {
	if rd == nil {
		return errors.New("rd is nil") // Ensure the reader is not nil.
	}

	// Parse the tag header to get the version and size of the frames.
	header, err := parseHeader(rd)
	if errors.Is(err, ErrNoTag) || errors.Is(err, io.EOF) {
		// If there's no tag or EOF, initialize an empty tag with default settings.
		tag.init(rd, 0, 4)

		return nil
	}

	if err != nil {
		return fmt.Errorf("error by parsing tag header: %w", err)
	}

	// Only ID3v2.3 and ID3v2.4 are supported.
	if header.Version < 3 {
		return ErrUnsupportedVersion
	}

	// Initialize the tag with the parsed header information.
	tag.init(rd, tagHeaderSize+header.FramesSize, header.Version)

	// If parsing is disabled, return early.
	if !opts.Parse {
		return nil
	}

	// Parse the frames within the tag.
	return tag.parseFrames(opts)
}

// init initializes the tag with the provided reader, size, and version.
// It also sets the default encoding based on the ID3v2 version.
func (tag *Tag) init(rd io.Reader, originalSize int64, version byte) {
	tag.DeleteAllFrames() // Clear any existing frames.

	tag.reader = rd
	tag.originalSize = originalSize
	tag.version = version
	tag.setDefaultEncodingBasedOnVersion(version) // Set encoding based on version.
}

// parseFrames parses the frames within the tag according to the provided options.
func (tag *Tag) parseFrames(opts Options) error {
	framesSize := tag.originalSize - tagHeaderSize // Calculate the remaining size for frames.

	// Create a map of frame IDs to parse based on the provided options.
	parseableIDs := tag.makeIDsFromDescriptions(opts.ParseFrames)
	isParseFramesProvided := len(opts.ParseFrames) > 0

	// Determine if the tag uses synch-safe sizes (ID3v2.4 feature).
	synchSafe := tag.Version() == 4

	// Get a buffered reader and a reusable byte slice for parsing.
	br := getBufReader(nil)
	defer putBufReader(br)

	buf := getByteSlice(defaultBufferSize)
	defer putByteSlice(buf)

	// Iterate through the frames until the remaining size is exhausted.
	for framesSize > 0 {
		header, err := parseFrameHeader(buf, tag.reader, synchSafe)
		if errors.Is(err, io.EOF) || errors.Is(err, ErrBlankFrame) || errors.Is(err, ErrInvalidSizeFormat) {
			break // Stop parsing if we hit EOF or encounter an invalid frame.
		}

		if err != nil {
			return err
		}

		id, bodySize := header.ID, header.BodySize

		// Update the remaining size after accounting for the current frame.
		framesSize -= frameHeaderSize + bodySize
		if framesSize < 0 {
			return ErrBodyOverflow // Frame exceeds the remaining tag size.
		}

		// Create a limited reader for the frame's body.
		bodyReader := getLimitedReader(tag.reader, bodySize)
		defer putLimitedReader(bodyReader)

		// Skip frames that are not in the list of frames to parse.
		if isParseFramesProvided && !parseableIDs[id] {
			if err = skipReaderBuf(bodyReader, buf); err != nil {
				return err
			}

			continue
		}

		// Reset the buffered reader to read the frame's body.
		br.Reset(bodyReader)

		// Parse the frame's body based on its ID.
		frame, err := parseFrameBody(id, br, tag.version)
		if err != nil && !errors.Is(err, io.EOF) {
			return err
		}

		// Add the parsed frame to the tag.
		tag.AddFrame(id, frame)

		// If parsing specific frames and this frame is not part of a sequence,
		// remove it from the list of frames to parse.
		if isParseFramesProvided && !mustFrameBeInSequence(id) {
			delete(parseableIDs, id)

			// If no more frames to parse, stop.
			if len(parseableIDs) == 0 {
				break
			}
		}

		if errors.Is(err, io.EOF) {
			break // Stop parsing if we hit EOF.
		}
	}

	return nil
}

// makeIDsFromDescriptions converts a list of frame descriptions into a map of frame IDs.
func (tag *Tag) makeIDsFromDescriptions(parseFrames []string) map[string]bool {
	ids := make(map[string]bool, len(parseFrames))

	for _, description := range parseFrames {
		ids[tag.CommonID(description)] = true // Map descriptions to their corresponding IDs.
	}

	return ids
}

// parseFrameHeader reads and parses the header of an ID3v2 frame.
func parseFrameHeader(buf []byte, rd io.Reader, synchSafe bool) (frameHeader, error) {
	var header frameHeader

	if len(buf) < frameHeaderSize {
		return header, errors.New("parseFrameHeader: buf is smaller than frame header size")
	}

	// Read the frame header into the buffer.
	fhBuf := buf[:frameHeaderSize]
	if _, err := rd.Read(fhBuf); err != nil {
		return header, err
	}

	id := fhBuf[:4] // Extract the frame ID.

	// Parse the frame's body size, considering synch-safe encoding if necessary.
	bodySize, err := parseSize(fhBuf[4:8], synchSafe)
	if err != nil {
		return header, err
	}

	// Check if the frame ID or size is invalid.
	if bytes.Equal(id, []byte{0, 0, 0, 0}) || bodySize == 0 {
		return header, ErrBlankFrame
	}

	header.ID = string(id)
	header.BodySize = bodySize

	return header, nil
}

// skipReaderBuf reads and discards data from the reader until EOF.
func skipReaderBuf(rd io.Reader, buf []byte) error {
	for {
		_, err := rd.Read(buf)
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}
	}

	return nil
}

// parseFrameBody parses the body of a frame based on its ID.
func parseFrameBody(id string, br *bufferedReader, version byte) (Framer, error) {
	// Handle text frames (frames starting with 'T').
	if id[0] == 'T' && id != UserDefinedTextFrameID {
		return parseTextFrame(br)
	}

	// Use the appropriate parser for known frame types.
	if parseFunc, exists := parsers[id]; exists {
		return parseFunc(br, version)
	}

	// Fall back to parsing unknown frames.
	return parseUnknownFrame(br)
}
