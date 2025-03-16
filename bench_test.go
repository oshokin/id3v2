package id3v2

import (
	"bytes"
	"io"
	"testing"
)

var frontCoverPicture = mustReadFile(frontCoverPath)

func BenchmarkParseAllFrames(b *testing.B) {
	writeTag(b, EncodingUTF8)

	musicContent := mustReadFile(mp3Path)

	b.ResetTimer()

	for range b.N {
		tag, err := ParseReader(bytes.NewReader(musicContent), parseOpts)
		if tag == nil || err != nil {
			b.Fatal("Error while opening mp3 file:", err)
		}
	}
}

func BenchmarkParseAllFramesISO(b *testing.B) {
	writeTag(b, EncodingISO)

	musicContent := mustReadFile(mp3Path)

	b.ResetTimer()

	for range b.N {
		tag, err := ParseReader(bytes.NewReader(musicContent), parseOpts)
		if tag == nil || err != nil {
			b.Fatal("Error while opening mp3 file:", err)
		}
	}
}

func BenchmarkParseArtistAndTitle(b *testing.B) {
	writeTag(b, EncodingUTF8)

	musicContent := mustReadFile(mp3Path)

	b.ResetTimer()

	for range b.N {
		opts := Options{Parse: true, ParseFrames: []string{ArtistFrameDescription, "Title"}}

		tag, err := ParseReader(bytes.NewReader(musicContent), opts)
		if tag == nil || err != nil {
			b.Fatal("Error while opening mp3 file:", err)
		}
	}
}

func BenchmarkWrite(b *testing.B) {
	for range b.N {
		benchWrite(b, EncodingUTF8)
	}
}

func BenchmarkWriteISO(b *testing.B) {
	for range b.N {
		benchWrite(b, EncodingISO)
	}
}

func benchWrite(b *testing.B, encoding Encoding) {
	tag := NewEmptyTag()
	setFrames(tag, encoding)

	if _, err := tag.WriteTo(io.Discard); err != nil {
		b.Error("Error while writing a tag:", err)
	}
}

func writeTag(b *testing.B, encoding Encoding) {
	tag, err := Open(mp3Path, Options{Parse: false})
	if tag == nil || err != nil {
		b.Fatal("Error while opening mp3 file:", err)
	}
	defer tag.Close()

	setFrames(tag, encoding)

	if err = tag.Save(); err != nil {
		b.Error("Error while saving a tag:", err)
	}
}

func setFrames(tag *Tag, encoding Encoding) {
	tag.SetTitle("Title")
	tag.SetArtist(ArtistFrameDescription)
	tag.SetAlbum("Album")
	tag.SetYear("2016")
	tag.SetGenre("Genre")

	pic := PictureFrame{
		Encoding:    encoding,
		MimeType:    "image/jpeg",
		PictureType: PTFrontCover,
		Description: "Front cover",
		Picture:     frontCoverPicture,
	}
	tag.AddAttachedPicture(pic)

	uslt := UnsynchronisedLyricsFrame{
		Encoding:          encoding,
		Language:          EnglishISO6392Code,
		ContentDescriptor: "Content descriptor",
		Lyrics:            "bogem/id3v2",
	}
	tag.AddUnsynchronisedLyricsFrame(uslt)

	comm := CommentFrame{
		Encoding:    encoding,
		Language:    EnglishISO6392Code,
		Description: "Short description",
		Text:        "The actual text",
	}
	tag.AddCommentFrame(comm)
}
