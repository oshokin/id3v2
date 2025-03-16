package id3v2

import (
	"encoding/binary"
	"errors"
	"io"
	"time"
)

const (
	// IgnoredOffset is a special value indicating that an offset should be ignored.
	IgnoredOffset = 0xFFFFFFFF

	// Number of nanoseconds in a millisecond.
	nanosInMillis = 1000000
)

// ChapterFrame represents a chapter frame in an ID3v2 tag,
// as defined by the ID3v2 chapters specification here - // according to spec from http://id3.org/id3v2-chapters-1.0.
// It supports a single TIT2 subframe (Title field) and ignores other subframes.
// If StartOffset or EndOffset equals IgnoredOffset,
// the corresponding time (StartTime or EndTime) should be used instead.
type ChapterFrame struct {
	ElementID   string        // Unique identifier for the chapter.
	StartTime   time.Duration // Start time of the chapter.
	EndTime     time.Duration // End time of the chapter.
	StartOffset uint32        // Start offset in bytes (optional, use IgnoredOffset to ignore).
	EndOffset   uint32        // End offset in bytes (optional, use IgnoredOffset to ignore).
	Title       *TextFrame    // Title of the chapter (optional).
	Description *TextFrame    // Description of the chapter (optional).
	Link        *LinkFrame    // Link associated with the chapter (optional).
	Artwork     *PictureFrame // Artwork associated with the chapter (optional).
}

// Size calculates the total size of the ChapterFrame in bytes, including all its subframes.
func (cf ChapterFrame) Size() int {
	size := encodedSize(cf.ElementID, EncodingISO) +
		1 + // Trailing zero after ElementID.
		4 + 4 + 4 + 4 // Sizes for StartTime, EndTime, StartOffset, and EndOffset.

	if cf.Title != nil {
		size += frameHeaderSize + cf.Title.Size() // Add size of the Title frame.
	}

	if cf.Description != nil {
		size += frameHeaderSize + cf.Description.Size() // Add size of the Description frame.
	}

	return size
}

// UniqueIdentifier returns the unique identifier for the ChapterFrame, which is its ElementID.
func (cf ChapterFrame) UniqueIdentifier() string {
	return cf.ElementID
}

// WriteTo writes the ChapterFrame to the provided io.Writer, including all its subframes.
func (cf ChapterFrame) WriteTo(w io.Writer) (n int64, err error) {
	return useBufferedWriter(w, func(bw *bufferedWriter) error {
		// Write the ElementID in ISO encoding, followed by a null terminator.
		bw.EncodeAndWriteText(cf.ElementID, EncodingISO)
		bw.WriteByte(0)

		// Write StartTime and EndTime in milliseconds, converting from nanoseconds.
		err = binary.Write(bw, binary.BigEndian, truncateInt64ToInt32(int64(cf.StartTime/nanosInMillis)))
		if err != nil {
			return err
		}

		err = binary.Write(bw, binary.BigEndian, truncateInt64ToInt32(int64(cf.EndTime/nanosInMillis)))
		if err != nil {
			return err
		}

		// Write StartOffset and EndOffset.
		err = binary.Write(bw, binary.BigEndian, cf.StartOffset)
		if err != nil {
			return err
		}

		err = binary.Write(bw, binary.BigEndian, cf.EndOffset)
		if err != nil {
			return err
		}

		// Write the Title frame if it exists.
		if cf.Title != nil {
			err = writeFrame(bw, TitleFrameID, *cf.Title, true)
			if err != nil {
				return err
			}
		}

		// Write the Description frame if it exists.
		if cf.Description != nil {
			err = writeFrame(bw, SubtitleRefinementFrameID, *cf.Description, true)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

// parseChapterFrame parses a ChapterFrame from a bufferedReader.
func parseChapterFrame(br *bufferedReader, version byte) (Framer, error) {
	var (
		elementID   = br.ReadText(EncodingISO) // Read the ElementID.
		synchSafe   = version == 4             // Determine if the frame uses synch-safe encoding.
		startTime   uint32
		startOffset uint32
		endTime     uint32
		endOffset   uint32
	)

	// Read StartTime, EndTime, StartOffset, and EndOffset from the buffer.
	if err := binary.Read(br, binary.BigEndian, &startTime); err != nil {
		return nil, err
	}

	if err := binary.Read(br, binary.BigEndian, &endTime); err != nil {
		return nil, err
	}

	if err := binary.Read(br, binary.BigEndian, &startOffset); err != nil {
		return nil, err
	}

	if err := binary.Read(br, binary.BigEndian, &endOffset); err != nil {
		return nil, err
	}

	var (
		title       TextFrame
		description TextFrame
		link        LinkFrame
		artwork     PictureFrame
		buf         = getByteSlice(defaultBufferSize)
	)

	defer putByteSlice(buf) // Return the buffer to the pool when done.

	// Parse subframes until the end of the chapter frame.
	for {
		header, err := parseFrameHeader(buf, br, synchSafe)
		if errors.Is(err, io.EOF) || errors.Is(err, ErrBlankFrame) || errors.Is(err, ErrInvalidSizeFormat) {
			break // Stop parsing if we reach the end or encounter an invalid frame.
		}

		if err != nil {
			return nil, err
		}

		id, bodySize := header.ID, header.BodySize

		// Handle Title and Description subframes.
		if id == TitleFrameID || id == SubtitleRefinementFrameID {
			bodyReader := getLimitedReader(br, bodySize)
			frameReaderReader := newBufferedReader(bodyReader)

			var frame Framer

			frame, err = parseTextFrame(frameReaderReader)
			if err != nil {
				putLimitedReader(bodyReader)

				return nil, err
			}

			if id == TitleFrameID {
				title, _ = frame.(TextFrame)
			} else if id == SubtitleRefinementFrameID {
				description, _ = frame.(TextFrame)
			}

			putLimitedReader(bodyReader)
		}

		// Handle Link subframes.
		if id == "WXXX" {
			bodyReader := getLimitedReader(br, bodySize)
			br = newBufferedReader(bodyReader)

			//nolint:govet // Shadowing is not an issue here since we return on error.
			frame, err := parseLinkFrame(br)
			if err != nil {
				putLimitedReader(bodyReader)

				return nil, err
			}

			link, _ = frame.(LinkFrame)

			putLimitedReader(bodyReader)
		}

		// Handle Artwork subframes.
		if id == "APIC" {
			bodyReader := getLimitedReader(br, bodySize)
			br = newBufferedReader(bodyReader)

			//nolint:govet // Shadowing is not an issue here since we return on error.
			frame, err := parsePictureFrame(br, version)
			if err != nil {
				putLimitedReader(bodyReader)

				return nil, err
			}

			artwork, _ = frame.(PictureFrame)

			putLimitedReader(bodyReader)
		}
	}

	// Construct and return the ChapterFrame.
	cf := ChapterFrame{
		ElementID:   string(elementID),
		StartTime:   time.Duration(int64(startTime) * nanosInMillis), // Convert milliseconds to nanoseconds.
		EndTime:     time.Duration(int64(endTime) * nanosInMillis),
		StartOffset: startOffset,
		EndOffset:   endOffset,
		Title:       &title,
		Description: &description,
		Link:        &link,
		Artwork:     &artwork,
	}

	return cf, nil
}
