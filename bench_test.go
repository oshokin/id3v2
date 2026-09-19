package id3v2

import (
	"bytes"
	"io"
	"testing"
)

var frontCoverPicture = mustReadFile(frontCoverPath)

func BenchmarkParseAllFrames(b *testing.B) {
	musicContent := mustReadFile(writeTag(b, EncodingUTF8))

	b.ResetTimer()

	for b.Loop() {
		tag, err := ParseReader(bytes.NewReader(musicContent), parseOpts)
		if tag == nil || err != nil {
			b.Fatal("Error while opening mp3 file:", err)
		}
	}
}

func BenchmarkParseAllFramesISO(b *testing.B) {
	musicContent := mustReadFile(writeTag(b, EncodingISO))
	b.ResetTimer()

	for b.Loop() {
		tag, err := ParseReader(bytes.NewReader(musicContent), parseOpts)
		if tag == nil || err != nil {
			b.Fatal("Error while opening mp3 file:", err)
		}
	}
}

func BenchmarkParseArtistAndTitle(b *testing.B) {
	musicContent := mustReadFile(writeTag(b, EncodingUTF8))

	b.ResetTimer()

	for b.Loop() {
		opts := Options{Parse: true, ParseFrames: []string{ArtistFrameDescription, "Title"}}

		tag, err := ParseReader(bytes.NewReader(musicContent), opts)
		if tag == nil || err != nil {
			b.Fatal("Error while opening mp3 file:", err)
		}
	}
}

func BenchmarkWrite(b *testing.B) {
	for b.Loop() {
		benchWrite(b, EncodingUTF8)
	}
}

func BenchmarkWriteISO(b *testing.B) {
	for b.Loop() {
		benchWrite(b, EncodingISO)
	}
}

func BenchmarkParseTag(b *testing.B) {
	content := mustReadFile(writeTag(b, EncodingUTF8))
	b.ResetTimer()

	for b.Loop() {
		tag, err := ParseReader(bytes.NewReader(content), parseOpts)
		if err != nil {
			b.Fatal(err)
		}

		_ = tag.Close()
	}
}

func BenchmarkWriteTag(b *testing.B) {
	for b.Loop() {
		benchWrite(b, EncodingUTF8)
	}
}

func BenchmarkParseAndWriteTag(b *testing.B) {
	content := mustReadFile(writeTag(b, EncodingUTF8))
	b.ResetTimer()

	for b.Loop() {
		tag, err := ParseReader(bytes.NewReader(content), parseOpts)
		if err != nil {
			b.Fatal(err)
		}

		if _, err := tag.WriteTo(io.Discard); err != nil {
			b.Fatal(err)
		}

		_ = tag.Close()
	}
}

func benchWrite(b *testing.B, encoding Encoding) {
	tag := NewEmptyTag()
	setFrames(tag, encoding)

	if _, err := tag.WriteTo(io.Discard); err != nil {
		b.Error("Error while writing a tag:", err)
	}
}

func writeTag(b *testing.B, encoding Encoding) string {
	path := copyFixture(b)

	tag, err := Open(path, Options{Parse: false})
	if tag == nil || err != nil {
		b.Fatal("Error while opening mp3 file:", err)
	}
	defer tag.Close()

	setFrames(tag, encoding)

	if err = tag.Save(); err != nil {
		b.Fatal("Error while saving a tag:", err)
	}

	return path
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
