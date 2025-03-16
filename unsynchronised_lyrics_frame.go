package id3v2

import "io"

// UnsynchronisedLyricsFrame represents an ID3v2 unsynchronized lyrics/text frame (USLT).
// This frame is used to store lyrics or text that is not synchronized with the audio.
// For example, it can be used to store the full lyrics of a song.
//
// To add this frame to a tag, use the `tag.AddUnsynchronisedLyricsFrame` method.
//
// The `Language` field must be a valid three-letter language code from the ISO 639-2 standard.
// You can find the list of valid codes here: https://www.loc.gov/standards/iso639-2/php/code_list.php
type UnsynchronisedLyricsFrame struct {
	Encoding          Encoding // The text encoding used for the lyrics and content descriptor.
	Language          string   // The language of the lyrics (e.g., "eng" for English).
	ContentDescriptor string   // A short description of the lyrics (e.g., "Verse 1").
	Lyrics            string   // The actual lyrics or text content.
}

// Size calculates the total size of the UnsynchronisedLyricsFrame in bytes.
// This includes the encoding byte, language code, content descriptor, and lyrics,
// as well as the termination bytes required by the encoding.
func (uslf UnsynchronisedLyricsFrame) Size() int {
	return 1 + // Encoding byte
		len(uslf.Language) + // Language code (always 3 bytes)
		encodedSize(uslf.ContentDescriptor, uslf.Encoding) + // Content descriptor size
		len(uslf.Encoding.TerminationBytes) + // Termination bytes for the descriptor
		encodedSize(uslf.Lyrics, uslf.Encoding) // Lyrics size
}

// UniqueIdentifier returns a string that uniquely identifies this frame.
// For UnsynchronisedLyricsFrame, the identifier is a combination of the language
// and content descriptor. This ensures that frames with the same language and
// descriptor are treated as the same frame.
func (uslf UnsynchronisedLyricsFrame) UniqueIdentifier() string {
	return uslf.Language + uslf.ContentDescriptor
}

// WriteTo writes the UnsynchronisedLyricsFrame to the provided io.Writer.
// It returns the number of bytes written and any error encountered during the write.
// If the language code is not exactly 3 characters long, it returns ErrInvalidLanguageLength.
func (uslf UnsynchronisedLyricsFrame) WriteTo(w io.Writer) (n int64, err error) {
	// Validate the language code length.
	if len(uslf.Language) != 3 {
		return n, ErrInvalidLanguageLength
	}

	// Use a buffered writer for efficient writing.
	return useBufferedWriter(w, func(bw *bufferedWriter) error {
		// Write the encoding byte.
		bw.WriteByte(uslf.Encoding.Key)

		// Write the 3-character language code.
		bw.WriteString(uslf.Language)

		// Write the content descriptor, encoded according to the frame's encoding.
		bw.EncodeAndWriteText(uslf.ContentDescriptor, uslf.Encoding)

		// Write the termination bytes for the content descriptor.
		_, err = bw.Write(uslf.Encoding.TerminationBytes)
		if err != nil {
			return err
		}

		// Write the lyrics, encoded according to the frame's encoding.
		bw.EncodeAndWriteText(uslf.Lyrics, uslf.Encoding)

		return nil
	})
}

// parseUnsynchronisedLyricsFrame parses an UnsynchronisedLyricsFrame from a bufferedReader.
// It reads the encoding, language, content descriptor, and lyrics from the reader.
// If any error occurs during reading, it returns the error.
func parseUnsynchronisedLyricsFrame(br *bufferedReader, _ byte) (Framer, error) {
	// Read the encoding byte and resolve the encoding type.
	encoding := getEncoding(br.ReadByte())

	// Read the 3-character language code.
	language := br.Next(3)

	// Read the content descriptor, using the frame's encoding.
	contentDescriptor := br.ReadText(encoding)

	// Check for errors after reading the descriptor.
	if br.Err() != nil {
		return nil, br.Err()
	}

	// Use a bytes.Buffer to read the lyrics.
	lyrics := getBytesBuffer()
	defer putBytesBuffer(lyrics) // Ensure the buffer is returned to the pool.

	// Read the lyrics from the reader into the buffer.
	if _, err := lyrics.ReadFrom(br); err != nil {
		return nil, err
	}

	// Create and return the UnsynchronisedLyricsFrame.
	uslf := UnsynchronisedLyricsFrame{
		Encoding:          encoding,
		Language:          string(language),
		ContentDescriptor: decodeText(contentDescriptor, encoding),
		Lyrics:            decodeText(lyrics.Bytes(), encoding),
	}

	return uslf, nil
}
