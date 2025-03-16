package id3v2

import "io"

// CommentFrame represents a comment frame in an ID3v2 tag,
// commonly used for storing comments about the audio file.
// Each comment frame includes a language code, a description, and the actual comment text.
// The language code must be a valid three-letter ISO 639-2 code.
// ISO 639-2 code list: https://www.loc.gov/standards/iso639-2/php/code_list.php.
//
// To add a comment frame to a tag, use the `tag.AddCommentFrame()` method.
type CommentFrame struct {
	Encoding    Encoding // The text encoding used for the description and comment text.
	Language    string   // A three-letter language code (e.g., "eng" for English).
	Description string   // A short description of the comment (e.g., "Recording Info").
	Text        string   // The actual comment text.
}

// Size calculates the total size of the comment frame in bytes, including the encoding byte,
// language code, description, termination bytes, and comment text.
func (cf CommentFrame) Size() int {
	return 1 + // Encoding byte
		len(cf.Language) + // Language code (always 3 bytes)
		encodedSize(cf.Description, cf.Encoding) + // Size of the encoded description
		len(cf.Encoding.TerminationBytes) + // Size of the termination bytes
		encodedSize(cf.Text, cf.Encoding) // Size of the encoded comment text
}

// UniqueIdentifier returns a string that uniquely identifies this comment frame.
// It combines the language code and description to ensure uniqueness,
// as multiple comment frames can exist in a tag as long as their language and description differ.
func (cf CommentFrame) UniqueIdentifier() string {
	return cf.Language + cf.Description
}

// WriteTo writes the comment frame to the provided io.Writer.
// It returns the number of bytes written and any error encountered during the write operation.
func (cf CommentFrame) WriteTo(w io.Writer) (n int64, err error) {
	// Ensure the language code is exactly 3 characters long, as required by the ID3v2 spec.
	if len(cf.Language) != 3 {
		return n, ErrInvalidLanguageLength
	}

	// Use a buffered writer for efficient writing.
	return useBufferedWriter(w, func(bw *bufferedWriter) error {
		// Write the encoding byte.
		bw.WriteByte(cf.Encoding.Key)

		// Write the language code.
		bw.WriteString(cf.Language)

		// Write the encoded description.
		bw.EncodeAndWriteText(cf.Description, cf.Encoding)

		// Write the termination bytes for the description.
		_, err = bw.Write(cf.Encoding.TerminationBytes)
		if err != nil {
			return err
		}

		// Write the encoded comment text.
		bw.EncodeAndWriteText(cf.Text, cf.Encoding)

		return nil
	})
}

// parseCommentFrame reads a comment frame from a buffered reader and returns a CommentFrame struct.
func parseCommentFrame(br *bufferedReader, _ byte) (Framer, error) {
	// Read the encoding byte and determine the text encoding.
	encoding := getEncoding(br.ReadByte())

	// Read the next 3 bytes as the language code.
	language := br.Next(3)

	// Read the description text, which is encoded according to the specified encoding.
	description := br.ReadText(encoding)

	// Check for any errors encountered so far.
	if br.Err() != nil {
		return nil, br.Err()
	}

	// Use a buffer to read the remaining bytes as the comment text.
	text := getBytesBuffer()
	defer putBytesBuffer(text)

	// Read the comment text from the buffered reader.
	if _, err := text.ReadFrom(br); err != nil {
		return nil, err
	}

	// Decode the description and comment text using the specified encoding.
	cf := CommentFrame{
		Encoding:    encoding,
		Language:    string(language),
		Description: decodeText(description, encoding),
		Text:        decodeText(text.Bytes(), encoding),
	}

	return cf, nil
}
