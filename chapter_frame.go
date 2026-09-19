package id3v2

import (
	"encoding/binary"
	"errors"
	"fmt"
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
// according to spec from http://id3.org/id3v2-chapters-1.0.
// If StartOffset or EndOffset equals IgnoredOffset,
// the corresponding time (StartTime or EndTime) should be used instead.
// Nested sub-frames are serialized with the parent tag version:
// ID3v2.3 uses unsynchsafe sizes, ID3v2.4 uses synchsafe sizes.
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
	return cf.sizeForVersion(4)
}

func (cf ChapterFrame) sizeForVersion(version byte) int {
	size := encodedSize(cf.ElementID, EncodingISO) +
		1 + // Trailing zero after ElementID.
		4 + 4 + 4 + 4 // Sizes for StartTime, EndTime, StartOffset, and EndOffset.

	if cf.Title != nil {
		size += frameHeaderSize + frameSizeForVersion(*cf.Title, version)
	}

	if cf.Description != nil {
		size += frameHeaderSize + frameSizeForVersion(*cf.Description, version)
	}

	if cf.Link != nil {
		size += frameHeaderSize + frameSizeForVersion(*cf.Link, version)
	}

	if cf.Artwork != nil {
		size += frameHeaderSize + frameSizeForVersion(*cf.Artwork, version)
	}

	return size
}

// UniqueIdentifier returns the unique identifier for the ChapterFrame, which is its ElementID.
func (cf ChapterFrame) UniqueIdentifier() string {
	return cf.ElementID
}

// WriteTo writes the ChapterFrame to the provided io.Writer, including all its subframes.
func (cf ChapterFrame) WriteTo(w io.Writer) (int64, error) {
	return cf.writeToVersion(w, 4)
}

func (cf ChapterFrame) writeToVersion(w io.Writer, version byte) (int64, error) {
	return useBufferedWriter(w, func(bw *bufferedWriter) error {
		// Write the ElementID in ISO encoding, followed by a null terminator.
		bw.EncodeAndWriteText(cf.ElementID, EncodingISO)
		bw.writeByte(0)

		// Write StartTime and EndTime in milliseconds, converting from nanoseconds.

		if err := binary.Write(
			bw,
			binary.BigEndian,
			truncateInt64ToInt32(int64(cf.StartTime/nanosInMillis)),
		); err != nil {
			return err
		}

		if err := binary.Write(
			bw,
			binary.BigEndian,
			truncateInt64ToInt32(int64(cf.EndTime/nanosInMillis)),
		); err != nil {
			return err
		}

		// Write StartOffset and EndOffset.
		if err := binary.Write(bw, binary.BigEndian, cf.StartOffset); err != nil {
			return err
		}

		if err := binary.Write(bw, binary.BigEndian, cf.EndOffset); err != nil {
			return err
		}

		// Write optional embedded subframes.
		if cf.Title != nil {
			if err := writeEmbeddedFrame(bw, TitleFrameID, *cf.Title, version); err != nil {
				return err
			}
		}

		if cf.Description != nil {
			if err := writeEmbeddedFrame(bw, SubtitleRefinementFrameID, *cf.Description, version); err != nil {
				return err
			}
		}

		if cf.Link != nil {
			if err := writeEmbeddedFrame(bw, "WXXX", *cf.Link, version); err != nil {
				return err
			}
		}

		if cf.Artwork != nil {
			if err := writeEmbeddedFrame(bw, "APIC", *cf.Artwork, version); err != nil {
				return err
			}
		}

		return nil
	})
}

// parseChapterFrame parses a ChapterFrame from a bufferedReader.
func parseChapterFrame(br *bufferedReader, version byte) (Framer, error) {
	elementID := br.ReadText(EncodingISO) // Read the ElementID.
	synchSafe := version == 4             // Determine if nested frames use synch-safe encoding.

	var (
		startTime   uint32
		endTime     uint32
		startOffset uint32
		endOffset   uint32
	)

	if err := binary.Read(br, binary.BigEndian, &startTime); err != nil {
		return nil, fmt.Errorf("parse CHAP start time: %w", err)
	}

	if err := binary.Read(br, binary.BigEndian, &endTime); err != nil {
		return nil, fmt.Errorf("parse CHAP end time: %w", err)
	}

	if err := binary.Read(br, binary.BigEndian, &startOffset); err != nil {
		return nil, fmt.Errorf("parse CHAP start offset: %w", err)
	}

	if err := binary.Read(br, binary.BigEndian, &endOffset); err != nil {
		return nil, fmt.Errorf("parse CHAP end offset: %w", err)
	}

	var (
		title       *TextFrame
		description *TextFrame
		link        *LinkFrame
		artwork     *PictureFrame
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
			return nil, fmt.Errorf("parse CHAP embedded frame header: %w", err)
		}

		id, bodySize := header.ID, header.BodySize

		// Skip unknown embedded frames without interpreting their body as a header.
		if id != TitleFrameID && id != SubtitleRefinementFrameID && id != "WXXX" && id != "APIC" {
			if _, skipErr := io.CopyN(io.Discard, br, bodySize); skipErr != nil {
				return nil, fmt.Errorf("skip unknown CHAP embedded frame %q: %w", id, skipErr)
			}

			continue
		}

		embeddedBody := &io.LimitedReader{
			R: br,
			N: bodySize,
		}
		embeddedReader := newBufferedReader(embeddedBody)

		switch id {
		case TitleFrameID:
			frame, parseErr := parseTextFrame(embeddedReader)
			if parseErr != nil {
				return nil, fmt.Errorf("parse CHAP embedded frame %q: %w", id, parseErr)
			}

			parsed, ok := frame.(TextFrame)
			if !ok {
				return nil, fmt.Errorf("parse CHAP embedded frame %q: unexpected type %T", id, frame)
			}

			title = &parsed

		case SubtitleRefinementFrameID:
			frame, parseErr := parseTextFrame(embeddedReader)
			if parseErr != nil {
				return nil, fmt.Errorf("parse CHAP embedded frame %q: %w", id, parseErr)
			}

			parsed, ok := frame.(TextFrame)
			if !ok {
				return nil, fmt.Errorf("parse CHAP embedded frame %q: unexpected type %T", id, frame)
			}

			description = &parsed

		case "WXXX":
			frame, parseErr := parseLinkFrame(embeddedReader)
			if parseErr != nil {
				return nil, fmt.Errorf("parse CHAP embedded frame %q: %w", id, parseErr)
			}

			parsed, ok := frame.(LinkFrame)
			if !ok {
				return nil, fmt.Errorf("parse CHAP embedded frame %q: unexpected type %T", id, frame)
			}

			link = &parsed

		case "APIC":
			frame, parseErr := parsePictureFrame(embeddedReader, version)
			if parseErr != nil {
				return nil, fmt.Errorf("parse CHAP embedded frame %q: %w", id, parseErr)
			}

			parsed, ok := frame.(PictureFrame)
			if !ok {
				return nil, fmt.Errorf("parse CHAP embedded frame %q: unexpected type %T", id, frame)
			}

			artwork = &parsed
		}

		if embeddedBody.N > 0 {
			if _, copyErr := io.Copy(io.Discard, embeddedBody); copyErr != nil {
				return nil, fmt.Errorf("drain CHAP embedded frame %q: %w", id, copyErr)
			}

			if embeddedBody.N > 0 {
				return nil, fmt.Errorf("parse CHAP embedded frame %q: truncated body", id)
			}
		}
	}

	// Construct and return the ChapterFrame.
	return ChapterFrame{
		ElementID:   string(elementID),
		StartTime:   time.Duration(int64(startTime) * nanosInMillis), // Convert milliseconds to nanoseconds.
		EndTime:     time.Duration(int64(endTime) * nanosInMillis),
		StartOffset: startOffset,
		EndOffset:   endOffset,
		Title:       title,
		Description: description,
		Link:        link,
		Artwork:     artwork,
	}, nil
}
