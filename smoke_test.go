package id3v2

import (
	"bytes"
	"testing"
)

func TestZvukGrabberWriteWorkflow(t *testing.T) {
	path := copyFixture(t)

	tag, err := Open(path, Options{Parse: false})
	if err != nil {
		t.Fatalf("Open(Parse:false): %v", err)
	}

	if tag.HasFrames() {
		t.Fatal("Parse:false must not parse existing frames")
	}

	before := audioPayload(t, path, tag.originalSize)

	tag.SetDefaultEncoding(EncodingUTF8)
	tag.SetTitle("Title")
	tag.SetArtist("Artist")
	tag.SetAlbum("Album")
	tag.SetYear("2026")
	tag.SetGenre("Genre")
	tag.AddAttachedPicture(PictureFrame{
		Encoding:    EncodingUTF8,
		MimeType:    "image/jpeg",
		PictureType: PTFrontCover,
		Picture:     []byte{1, 2, 3},
	})
	tag.AddCommentFrame(CommentFrame{
		Encoding: EncodingUTF8,
		Language: EnglishISO6392Code,
		Text:     "comment",
	})
	tag.AddUnsynchronisedLyricsFrame(UnsynchronisedLyricsFrame{
		Encoding: EncodingUTF8,
		Language: EnglishISO6392Code,
		Lyrics:   "lyrics",
	})
	tag.AddSynchronisedLyricsFrame(SynchronisedLyricsFrame{
		Encoding:        EncodingUTF8,
		Language:        EnglishISO6392Code,
		TimestampFormat: SYLTAbsoluteMillisecondsTimestampFormat,
		ContentType:     SYLTLyricsContentType,
		SynchronizedTexts: []SynchronizedText{
			{Text: "line", Timestamp: 100},
		},
	})

	if err := tag.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	if err := tag.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	reopened, err := Open(path, Options{Parse: true})
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer reopened.Close()

	if reopened.Title() != "Title" || reopened.Artist() != "Artist" ||
		reopened.Album() != "Album" || reopened.Year() != "2026" || reopened.Genre() != "Genre" {
		t.Fatalf("text frames not preserved: %q %q %q %q %q",
			reopened.Title(), reopened.Artist(), reopened.Album(), reopened.Year(), reopened.Genre())
	}

	after := audioPayload(t, path, reopened.originalSize)
	if !bytes.Equal(before, after) {
		t.Fatal("audio payload changed")
	}
}

func TestPublicCompositeLiteralsCompile(t *testing.T) {
	t.Parallel()

	_ = Options{Parse: false}
	_ = PictureFrame{
		Encoding:    EncodingUTF8,
		MimeType:    "image/jpeg",
		PictureType: PTFrontCover,
		Picture:     []byte{},
	}
	_ = CommentFrame{
		Encoding: EncodingUTF8,
		Language: EnglishISO6392Code,
	}
	_ = SynchronisedLyricsFrame{
		Encoding:        EncodingUTF8,
		Language:        EnglishISO6392Code,
		TimestampFormat: SYLTAbsoluteMillisecondsTimestampFormat,
		ContentType:     SYLTLyricsContentType,
	}
}
