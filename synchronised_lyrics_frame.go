package id3v2

import (
	"encoding/binary"
	"errors"
	"io"
	"regexp"
	"strconv"
	"strings"
)

type (
	// SYLTTimestampFormat represents the format used for timestamps in a SYLT frames.
	SYLTTimestampFormat byte

	// SYLTContentType represents the type of synchronized content in a SYLT frame.
	SYLTContentType byte

	// SynchronisedLyricsFrame represents a SYLT (Synchronised Lyrics) frame in an ID3v2 tag.
	// This frame is used to store lyrics or text that are synchronized to specific timestamps in the audio.
	// For example, you can use this to display karaoke lyrics or subtitles that match the timing of the song.
	//
	// To add a SYLT frame to a tag, use the `tag.AddSynchronisedLyricsFrame` function.
	//
	// The `Language` field must be a valid three-letter language code from the ISO 639-2 standard.
	// You can find the list of codes here: https://www.loc.gov/standards/iso639-2/php/code_list.php
	SynchronisedLyricsFrame struct {
		Encoding          Encoding            // The text encoding used for the lyrics and descriptor.
		Language          string              // The language of the lyrics (e.g., "eng" for English).
		TimestampFormat   SYLTTimestampFormat // The format of the timestamps (e.g., milliseconds or MPEG frames).
		ContentType       SYLTContentType     // The type of content (e.g., lyrics, transcription, or events).
		ContentDescriptor string              // A description of the content (e.g., "Verse 1").
		SynchronizedTexts []SynchronizedText  // A list of synchronized text entries with their timestamps.
	}

	// ParseLRCFileParsingResult holds the result of parsing an LRC file.
	ParseLRCFileParsingResult struct {
		TimestampFormat   SYLTTimestampFormat // The format of the timestamps in the parsed file.
		Metadata          map[string]string   // Metadata extracted from the LRC file.
		SynchronizedTexts []SynchronizedText  // A list of synchronized text entries with their timestamps.
		Comments          map[int]string      // Comments extracted from the LRC file, keyed by line number.
	}

	// SynchronizedText represents a single synchronized text entry with its associated timestamp.
	SynchronizedText struct {
		Text      string // The text to display (e.g., a line of lyrics).
		Timestamp uint32 // The timestamp or frame number in the audio when the text should be displayed.
	}
)

// Constants for the timestamp format in a SYLT frame.
const (
	SYLTUnknownTimestampFormat              SYLTTimestampFormat = iota // Unknown timestamp format.
	SYLTAbsoluteMpegFramesTimestampFormat                              // Timestamps are in MPEG frames.
	SYLTAbsoluteMillisecondsTimestampFormat                            // Timestamps are in milliseconds.
)

// Constants for the content type in a SYLT frame.
const (
	SYLTOtherContentType             SYLTContentType = iota // The content type is unspecified or other.
	SYLTLyricsContentType                                   // The content is lyrics.
	SYLTTextTranscriptionContentType                        // The content is a text transcription.
	SYLTMovementContentType                                 // The content describes movements (e.g., dance steps).
	SYLTEventsContentType                                   // The content describes events (e.g., sound effects).
	SYLTChordContentType                                    // The content is chord information.
	SYLTTriviaContentType                                   // The content is trivia or additional information.
	SYLTWebpageURLsContentType                              // The content contains URLs to webpages.
	SYLTImageURLsContentType                                // The content contains URLs to images.
)

// Metadata tag types for LRC files.
const (
	LRCTagTitle    = "ti"     // Title of the song.
	LRCTagArtist   = "ar"     // Artist performing the song.
	LRCTagAlbum    = "al"     // Album the song is from.
	LRCTagAuthor   = "au"     // Author of the song.
	LRCTagLyricist = "lr"     // Lyricist of the song.
	LRCTagLength   = "length" // Length of the song (mm:ss).
	LRCTagBy       = "by"     // Author of the LRC file (not the song).
	LRCTagOffset   = "offset" // Global offset value for the lyric times, in milliseconds.
	LRCTagTool     = "tool"   // The player or editor that created the LRC file.
	LRCTagVersion  = "ve"     // The version of the program.
)

var (
	// ContentType maps content type constants to their human-readable descriptions.
	ContentType = map[SYLTContentType]string{
		SYLTOtherContentType:             "Other",
		SYLTLyricsContentType:            "Lyrics",
		SYLTTextTranscriptionContentType: "Transcription",
		SYLTMovementContentType:          "Movement",
		SYLTEventsContentType:            "Events",
		SYLTChordContentType:             "Chord",
		SYLTTriviaContentType:            "Trivia",
		SYLTWebpageURLsContentType:       "WebpageUrls",
		SYLTImageURLsContentType:         "ImageUrls",
	}

	// SYLTMetadataPattern is a regex pattern to match metadata in LRC files (e.g., [key:value]).
	SYLTMetadataPattern = regexp.MustCompile(`^\[(\w+):(.+?)\]$`)

	// SYLTOffsetMetadataPattern is a regex pattern to match the offset metadata in LRC files (e.g., [offset:+500]).
	SYLTOffsetMetadataPattern = regexp.MustCompile(`^\[offset:([+-]?\d+)\]`)

	// SYLTTimestampPattern is a regex pattern to match timestamps in LRC files (e.g., [mm:ss.xx]).
	SYLTTimestampPattern = regexp.MustCompile(`\[(\d+):(\d{2})\.(\d{2})\](.*)`)
)

// Size calculates the total size of the SYLT frame in bytes.
func (sylf SynchronisedLyricsFrame) Size() int {
	var s int
	for _, v := range sylf.SynchronizedTexts {
		s += encodedSize(v.Text, sylf.Encoding)  // Size of the text.
		s += len(sylf.Encoding.TerminationBytes) // Size of the text termination bytes.
		s += 4                                   // Size of the timestamp (4 bytes for uint32).
	}

	// Add the size of the frame header fields.
	return 1 + // Encoding byte.
		len(sylf.Language) + // Language code (3 bytes).
		encodedSize(sylf.ContentDescriptor, sylf.Encoding) + // Size of the content descriptor.
		len(sylf.Encoding.TerminationBytes) + // Size of the descriptor termination bytes.
		s + // Size of all synchronized texts and timestamps.
		1 + // Timestamp format byte.
		1 // Content type byte.
}

// UniqueIdentifier returns a unique identifier for the SYLT frame.
func (sylf SynchronisedLyricsFrame) UniqueIdentifier() string {
	return sylf.Language + sylf.ContentDescriptor
}

// WriteTo writes the SYLT frame to the provided io.Writer.
func (sylf SynchronisedLyricsFrame) WriteTo(w io.Writer) (n int64, err error) {
	if len(sylf.Language) != 3 {
		return n, ErrInvalidLanguageLength // Ensure the language code is exactly 3 characters.
	}

	return useBufferedWriter(w, func(bw *bufferedWriter) error {
		// Write the frame header fields.
		bw.writeByte(sylf.Encoding.Key)                              // Write the encoding byte.
		bw.WriteString(sylf.Language)                                // Write the language code.
		bw.writeByte(byte(sylf.TimestampFormat))                     // Write the timestamp format.
		bw.writeByte(byte(sylf.ContentType))                         // Write the content type.
		bw.EncodeAndWriteText(sylf.ContentDescriptor, sylf.Encoding) // Write the content descriptor.

		// Write the descriptor termination bytes.
		_, err = bw.Write(sylf.Encoding.TerminationBytes)
		if err != nil {
			return err
		}

		// Write each synchronized text entry.
		for _, v := range sylf.SynchronizedTexts {
			bw.EncodeAndWriteText(v.Text, sylf.Encoding) // Write the text.

			_, err = bw.Write(sylf.Encoding.TerminationBytes) // Write the text termination bytes.
			if err != nil {
				return err
			}

			_, err = bw.Write(v.timestampToBigEndian()) // Write the timestamp.
			if err != nil {
				return err
			}
		}

		return nil
	})
}

// timestampToBigEndian converts a uint32 timestamp to a 4-byte slice in big-endian format.
func (sy SynchronizedText) timestampToBigEndian() []byte {
	bs := make([]byte, 4)
	binary.BigEndian.PutUint32(bs, sy.Timestamp)

	return bs
}

// ParseLRCFile reads and parses an LRC-formatted lyrics file from the provided io.Reader.
// It extracts synchronized lyrics, adjusts timestamps based on any offset, and returns the parsed result.
func ParseLRCFile(inputReader io.Reader) (ParseLRCFileParsingResult, error) {
	// Read and clean up lines from the input reader.
	lines, err := readLinesFromReader(inputReader,
		func(sourceLine string) (string, bool) {
			resultLine := strings.TrimSpace(sourceLine)
			isLineSkipped := resultLine == "" // Skip empty lines.

			return resultLine, isLineSkipped
		})
	if err != nil {
		return ParseLRCFileParsingResult{}, err
	}

	// Step 1: Check for an offset in the LRC file.
	offset := int64(0) // Default to no offset.

	// First pass: Look for an offset in the metadata.
	for _, line := range lines {
		match := SYLTOffsetMetadataPattern.FindStringSubmatch(line)
		if len(match) < 2 {
			continue // Skip lines that don't contain an offset.
		}

		var offsetValue int64

		offsetValue, err = strconv.ParseInt(match[1], 10, 64)
		if err != nil {
			return ParseLRCFileParsingResult{}, err // Return an error if the offset is invalid.
		}

		offset = offsetValue // Use the found offset.

		break // Stop searching after finding the first valid offset.
	}

	// Step 2: Parse the LRC file, adjusting timestamps using the offset (if any).
	result := ParseLRCFileParsingResult{
		// Assume timestamps are in milliseconds, as the LRC format uses [mm:ss.xx].
		TimestampFormat:   SYLTAbsoluteMillisecondsTimestampFormat,
		Metadata:          make(map[string]string),                 // Store metadata like artist or title.
		SynchronizedTexts: make([]SynchronizedText, 0, len(lines)), // Pre-allocate space for lyrics.
		Comments:          make(map[int]string),                    // Store comments from the LRC file.
	}

	// Second pass: Process each line to extract lyrics and metadata.
	for i, line := range lines {
		// Skip lines that contain the offset metadata (already processed).
		offsetMatch := SYLTOffsetMetadataPattern.FindStringSubmatch(line)
		if len(offsetMatch) > 0 {
			continue
		}

		// Check if the line contains metadata (e.g., [ar:Artist Name]).
		metadataMatch := SYLTMetadataPattern.FindStringSubmatch(line)
		// Check if the line contains a timestamp and lyrics (e.g., [01:23.45]Hello world).
		timestampMatch := SYLTTimestampPattern.FindStringSubmatch(line)

		switch {
		case len(timestampMatch) == 5:
			// Extract the timestamp components and lyrics.
			minutes, parseErr := strconv.ParseInt(timestampMatch[1], 10, 0)
			if parseErr != nil {
				return ParseLRCFileParsingResult{}, parseErr
			}

			seconds, parseErr := strconv.ParseInt(timestampMatch[2], 10, 0)
			if parseErr != nil {
				return ParseLRCFileParsingResult{}, parseErr
			}

			hundredths, parseErr := strconv.ParseInt(timestampMatch[3], 10, 0)
			if parseErr != nil {
				return ParseLRCFileParsingResult{}, parseErr
			}

			lyric := strings.TrimSpace(timestampMatch[4])

			// Convert the timestamp to milliseconds.
			timestamp := minutes*60*1000 + seconds*1000 + hundredths*10

			// Adjust the timestamp by the offset (if any).
			timestamp += offset

			// Add the synchronized lyrics to the result.
			result.SynchronizedTexts = append(result.SynchronizedTexts,
				SynchronizedText{
					Text:      lyric,
					Timestamp: truncateInt64ToUint32(timestamp),
				})
		case len(metadataMatch) == 3:
			// Store metadata key-value pairs (e.g., [ar:Artist Name] -> "ar": "Artist Name").
			result.Metadata[metadataMatch[1]] = metadataMatch[2]
		case strings.HasPrefix(line, "#"):
			// If the line starts with a '#', treat it as a comment.
			result.Comments[i+1] = strings.TrimPrefix(line, "#") // Store the comment with the line number as the key.
		}
	}

	return result, nil
}

// parseSynchronisedLyricsFrame parses a SYLT frame from a bufferedReader.
func parseSynchronisedLyricsFrame(br *bufferedReader, _ byte) (Framer, error) {
	encoding := getEncoding(br.readByte())     // Read the encoding byte.
	language := br.Next(3)                     // Read the language code.
	timestampFormat := br.readByte()           // Read the timestamp format.
	contentType := br.readByte()               // Read the content type.
	contentDescriptor := br.ReadText(encoding) // Read the content descriptor.

	if br.Err() != nil {
		return nil, br.Err() // Check for errors after reading.
	}

	var s []SynchronizedText

	for {
		var (
			textLyric []byte
			err       error
		)
		if encoding.Equals(EncodingUTF16) || encoding.Equals(EncodingUTF16BE) {
			textLyric, err = br.readTillUTF16Terminator()
		} else {
			textLyric, err = br.readTillDelimiters(encoding.TerminationBytes)
		}

		if err != nil {
			if errors.Is(err, io.EOF) && len(textLyric) == 0 {
				break
			}

			return nil, err
		}

		t := SynchronizedText{Text: decodeText(textLyric, encoding)}
		br.Next(len(encoding.TerminationBytes))

		if br.Err() != nil {
			return nil, br.Err()
		}

		timeStampUint, err := readUint32BE(br)
		if err != nil {
			return nil, err
		}

		t.Timestamp = timeStampUint
		s = append(s, t)
	}

	return SynchronisedLyricsFrame{
		Encoding:          encoding,
		Language:          string(language),
		TimestampFormat:   SYLTTimestampFormat(timestampFormat),
		ContentType:       SYLTContentType(contentType),
		ContentDescriptor: decodeText(contentDescriptor, encoding),
		SynchronizedTexts: s,
	}, nil
}
