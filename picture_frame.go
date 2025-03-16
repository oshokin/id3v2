package id3v2

import (
	"fmt"
	"io"
)

// PictureFrame represents an ID3v2 picture frame (APIC), which is used to store images like album art.
// To add a picture frame to a tag, use the `tag.AddAttachedPicture` method.
//
// The frame includes metadata like the MIME type, picture type (e.g., front cover, back cover),
// a description, and the actual image data.
type PictureFrame struct {
	Encoding    Encoding // The text encoding used for the description.
	MimeType    string   // The MIME type of the image (e.g., "image/jpeg").
	PictureType byte     // The type of picture (e.g., front cover, back cover).
	Description string   // A description of the picture.
	Picture     []byte   // The raw binary data of the image.
}

// UniqueIdentifier generates a unique string identifier for the PictureFrame.
// This is used to distinguish between multiple picture frames in a tag.
// The identifier combines the picture type and description.
func (pf PictureFrame) UniqueIdentifier() string {
	return fmt.Sprintf("%02X%s", pf.PictureType, pf.Description)
}

// Size calculates the total size of the PictureFrame in bytes.
// This includes the encoding byte, MIME type, picture type, description, and image data.
func (pf PictureFrame) Size() int {
	return 1 + // Encoding byte (1 byte for the encoding type)
		len(pf.MimeType) + // Length of the MIME type string (e.g., "image/jpeg")
		1 + // Null terminator for the MIME type string
		1 + // Picture type byte (1 byte for the type, e.g., front cover)
		encodedSize(pf.Description, pf.Encoding) + // Size of the encoded description
		len(pf.Encoding.TerminationBytes) + // Size of the termination bytes for the description
		len(pf.Picture) // Size of the raw image data
}

// WriteTo writes the PictureFrame to the provided io.Writer.
// It returns the number of bytes written and any error encountered.
// This method is used when saving the frame to an MP3 file.
func (pf PictureFrame) WriteTo(w io.Writer) (n int64, err error) {
	return useBufferedWriter(w, func(bw *bufferedWriter) error {
		// Write the encoding byte.
		bw.WriteByte(pf.Encoding.Key)

		// Write the MIME type and a null terminator.
		bw.WriteString(pf.MimeType)
		bw.WriteByte(0)

		// Write the picture type byte.
		bw.WriteByte(pf.PictureType)

		// Write the encoded description.
		bw.EncodeAndWriteText(pf.Description, pf.Encoding)

		// Write the termination bytes for the description.
		_, err = bw.Write(pf.Encoding.TerminationBytes)
		if err != nil {
			return err
		}

		// Write the raw image data.
		_, err = bw.Write(pf.Picture)
		if err != nil {
			return err
		}

		return nil
	})
}

// parsePictureFrame reads and parses a PictureFrame from a bufferedReader.
// It extracts the encoding, MIME type, picture type, description, and image data.
// This function is used when reading an MP3 file and decoding its ID3v2 tag.
func parsePictureFrame(br *bufferedReader, _ byte) (Framer, error) {
	// Read the encoding byte and determine the text encoding.
	encoding := getEncoding(br.ReadByte())

	// Read the MIME type as ISO-8859-1 encoded text.
	mimeType := br.ReadText(EncodingISO)

	// Read the picture type byte.
	pictureType := br.ReadByte()

	// Read the description using the specified encoding.
	description := br.ReadText(encoding)

	// Read the remaining bytes as the image data.
	picture := br.ReadAll()

	// Check for any errors during reading.
	if br.Err() != nil {
		return nil, br.Err()
	}

	// Create and return a PictureFrame with the parsed data.
	pf := PictureFrame{
		Encoding:    encoding,
		MimeType:    string(mimeType),
		PictureType: pictureType,
		Description: decodeText(description, encoding),
		Picture:     picture,
	}

	return pf, nil
}
