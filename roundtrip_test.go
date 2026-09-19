package id3v2

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math/big"
	"os"
	"testing"
	"testing/iotest"
	"time"
)

func TestTextFrameRoundTrip(t *testing.T) {
	t.Parallel()

	frame := TextFrame{Encoding: EncodingUTF8, Text: "Title"}
	assertFrameSizeMatchesWrite(t, frame)

	got := assertRoundTrip(t, 4, "TIT2", frame).(TextFrame)
	if got.Text != frame.Text {
		t.Fatalf("Text = %q, want %q", got.Text, frame.Text)
	}
}

func TestTextFrameV24MultiValueRoundTrip(t *testing.T) {
	t.Parallel()

	frame := TextFrame{
		Encoding: EncodingUTF8,
		Text:     "artist1",
		Multi:    []string{"artist1", "artist2"},
	}
	assertFrameSizeMatchesWrite(t, frame)
	assertVersionedFrameSizeMatchesWrite(t, frame, 4)

	got := assertRoundTrip(t, 4, "TPE1", frame).(TextFrame)
	if got.Text != "artist1" {
		t.Fatalf("Text = %q, want artist1", got.Text)
	}

	if len(got.Multi) != 2 || got.Multi[0] != "artist1" || got.Multi[1] != "artist2" {
		t.Fatalf("Multi = %#v, want [artist1 artist2]", got.Multi)
	}
}

func TestTextFrameV24UTF16MultiValueRoundTrip(t *testing.T) {
	t.Parallel()

	frame := TextFrame{
		Encoding: EncodingUTF16,
		Text:     "Ā",
		Multi:    []string{"Ā", "B"},
	}
	assertFrameSizeMatchesWrite(t, frame)
	assertVersionedFrameSizeMatchesWrite(t, frame, 4)

	got := assertRoundTrip(t, 4, "TPE1", frame).(TextFrame)
	if got.Text != "Ā" {
		t.Fatalf("Text = %q, want Ā", got.Text)
	}

	if len(got.Multi) != 2 || got.Multi[0] != "Ā" || got.Multi[1] != "B" {
		t.Fatalf("Multi = %#v, want [Ā B]", got.Multi)
	}
}

func TestTextFrameV24UTF16BEMultiValueRoundTrip(t *testing.T) {
	t.Parallel()

	frame := TextFrame{
		Encoding: EncodingUTF16BE,
		Text:     "Ā",
		Multi:    []string{"Ā", "B"},
	}
	assertFrameSizeMatchesWrite(t, frame)
	assertVersionedFrameSizeMatchesWrite(t, frame, 4)

	got := assertRoundTrip(t, 4, "TPE1", frame).(TextFrame)
	if got.Text != "Ā" {
		t.Fatalf("Text = %q, want Ā", got.Text)
	}

	if len(got.Multi) != 2 || got.Multi[0] != "Ā" || got.Multi[1] != "B" {
		t.Fatalf("Multi = %#v, want [Ā B]", got.Multi)
	}
}

func TestTextFrameV23PreservesLegacySingleValueBehavior(t *testing.T) {
	t.Parallel()

	frame := TextFrame{
		Encoding: EncodingISO,
		Text:     "artist1",
		Multi:    []string{"artist1", "artist2"},
	}
	assertVersionedFrameSizeMatchesWrite(t, frame, 3)

	got := assertRoundTrip(t, 3, "TPE1", frame).(TextFrame)
	if got.Text != "artist1" {
		t.Fatalf("Text = %q, want artist1", got.Text)
	}

	if len(got.Multi) != 1 {
		t.Fatalf("v2.3 Multi = %#v, want single value", got.Multi)
	}
}

func TestUserDefinedTextFrameRoundTrip(t *testing.T) {
	t.Parallel()

	frame := UserDefinedTextFrame{
		Encoding:    EncodingUTF8,
		Description: "MusicBrainz Album Id",
		Value:       "abc",
	}
	assertFrameSizeMatchesWrite(t, frame)

	got := assertRoundTrip(t, 4, UserDefinedTextFrameID, frame).(UserDefinedTextFrame)
	if got.Description != frame.Description || got.Value != frame.Value {
		t.Fatalf("got %#v", got)
	}
}

func TestUserDefinedTextFrameV24MultiValueRoundTrip(t *testing.T) {
	t.Parallel()

	frame := UserDefinedTextFrame{
		Encoding:    EncodingUTF8,
		Description: "multi",
		Value:       "val1",
		Multi:       []string{"val1", "val2"},
	}
	assertFrameSizeMatchesWrite(t, frame)
	assertVersionedFrameSizeMatchesWrite(t, frame, 4)

	got := assertRoundTrip(t, 4, UserDefinedTextFrameID, frame).(UserDefinedTextFrame)
	if len(got.Multi) != 2 || got.Multi[0] != "val1" || got.Multi[1] != "val2" {
		t.Fatalf("Multi = %#v", got.Multi)
	}
}

func TestUserDefinedTextFrameV24UTF16MultiValueRoundTrip(t *testing.T) {
	t.Parallel()

	frame := UserDefinedTextFrame{
		Encoding:    EncodingUTF16,
		Description: "multi",
		Value:       "val1",
		Multi:       []string{"val1", "val2"},
	}
	assertFrameSizeMatchesWrite(t, frame)
	assertVersionedFrameSizeMatchesWrite(t, frame, 4)

	got := assertRoundTrip(t, 4, UserDefinedTextFrameID, frame).(UserDefinedTextFrame)
	if len(got.Multi) != 2 || got.Multi[0] != "val1" || got.Multi[1] != "val2" {
		t.Fatalf("Multi = %#v", got.Multi)
	}
}

func TestCommentFrameRoundTrip(t *testing.T) {
	t.Parallel()

	frame := CommentFrame{
		Encoding:    EncodingUTF8,
		Language:    EnglishISO6392Code,
		Description: "desc",
		Text:        "comment",
	}
	assertFrameSizeMatchesWrite(t, frame)

	got := assertRoundTrip(t, 4, "COMM", frame).(CommentFrame)
	if got.Text != frame.Text || got.Language != frame.Language {
		t.Fatalf("got %#v", got)
	}
}

func TestPictureFrameRoundTrip(t *testing.T) {
	t.Parallel()

	frame := PictureFrame{
		Encoding:    EncodingUTF8,
		MimeType:    "image/jpeg",
		PictureType: PTFrontCover,
		Description: "cover",
		Picture:     []byte{1, 2, 3, 4},
	}
	assertFrameSizeMatchesWrite(t, frame)

	got := assertRoundTrip(t, 4, "APIC", frame).(PictureFrame)
	if !bytes.Equal(got.Picture, frame.Picture) || got.Description != frame.Description {
		t.Fatalf("got %#v", got)
	}
}

func TestUnsynchronisedLyricsFrameRoundTrip(t *testing.T) {
	t.Parallel()

	frame := UnsynchronisedLyricsFrame{
		Encoding:          EncodingUTF8,
		Language:          EnglishISO6392Code,
		ContentDescriptor: "desc",
		Lyrics:            "lyrics",
	}
	assertFrameSizeMatchesWrite(t, frame)

	got := assertRoundTrip(t, 4, "USLT", frame).(UnsynchronisedLyricsFrame)
	if got.Lyrics != frame.Lyrics {
		t.Fatalf("Lyrics = %q", got.Lyrics)
	}
}

func TestSynchronisedLyricsFrameRoundTrip(t *testing.T) {
	t.Parallel()

	frame := SynchronisedLyricsFrame{
		Encoding:          EncodingUTF8,
		Language:          EnglishISO6392Code,
		TimestampFormat:   SYLTAbsoluteMillisecondsTimestampFormat,
		ContentType:       SYLTLyricsContentType,
		ContentDescriptor: "Verse 1",
		SynchronizedTexts: []SynchronizedText{
			{Text: "First", Timestamp: 1000},
			{Text: "Second", Timestamp: 2000},
		},
	}
	assertFrameSizeMatchesWrite(t, frame)

	got := assertRoundTrip(t, 4, "SYLT", frame).(SynchronisedLyricsFrame)
	if len(got.SynchronizedTexts) != 2 {
		t.Fatalf("entries = %d", len(got.SynchronizedTexts))
	}

	if got.SynchronizedTexts[1].Timestamp != 2000 {
		t.Fatalf("timestamp = %d", got.SynchronizedTexts[1].Timestamp)
	}
}

func TestSynchronisedLyricsFrameTruncatedTimestamp(t *testing.T) {
	t.Parallel()

	frame := SynchronisedLyricsFrame{
		Encoding:          EncodingUTF8,
		Language:          EnglishISO6392Code,
		TimestampFormat:   SYLTAbsoluteMillisecondsTimestampFormat,
		ContentType:       SYLTLyricsContentType,
		ContentDescriptor: "x",
		SynchronizedTexts: []SynchronizedText{{Text: "line", Timestamp: 1}},
	}

	var buf bytes.Buffer
	if _, err := frame.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}

	truncated := buf.Bytes()[:buf.Len()-2]

	_, err := parseSynchronisedLyricsFrame(newBufferedReader(bytes.NewReader(truncated)), 4)
	if err == nil {
		t.Fatal("expected error for truncated SYLT timestamp")
	}
}

func TestUFIDFrameRoundTrip(t *testing.T) {
	t.Parallel()

	frame := UFIDFrame{
		OwnerIdentifier: "https://musicbrainz.org",
		Identifier:      []byte("id-123"),
	}
	assertFrameSizeMatchesWrite(t, frame)

	got := assertRoundTrip(t, 4, "UFID", frame).(UFIDFrame)
	if got.OwnerIdentifier != frame.OwnerIdentifier || !bytes.Equal(got.Identifier, frame.Identifier) {
		t.Fatalf("got %#v", got)
	}
}

func TestUFIDFrameRejectsEmptyOwner(t *testing.T) {
	t.Parallel()

	_, err := UFIDFrame{Identifier: []byte("x")}.WriteTo(&bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error for empty UFID owner")
	}
}

func TestUFIDFrameRejectsIdentifierLongerThan64Bytes(t *testing.T) {
	t.Parallel()

	_, err := UFIDFrame{
		OwnerIdentifier: "owner",
		Identifier:      bytes.Repeat([]byte("a"), 65),
	}.WriteTo(&bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error for oversized UFID identifier")
	}
}

func TestUFIDFrameSizeMatchesWriteTo(t *testing.T) {
	t.Parallel()

	assertFrameSizeMatchesWrite(t, UFIDFrame{
		OwnerIdentifier: "owner",
		Identifier:      []byte{1, 2, 3, 4},
	})
}

func TestPopularimeterFrameRoundTrip(t *testing.T) {
	t.Parallel()

	frame := PopularimeterFrame{
		Email:   "foo@bar.com",
		Rating:  128,
		Counter: big.NewInt(42),
	}
	assertFrameSizeMatchesWrite(t, frame)

	got := assertRoundTrip(t, 4, "POPM", frame).(PopularimeterFrame)
	if got.Email != frame.Email || got.Rating != frame.Rating || got.Counter.Cmp(frame.Counter) != 0 {
		t.Fatalf("got %#v", got)
	}
}

func TestPopularimeterFrameZeroValueDoesNotPanic(t *testing.T) {
	t.Parallel()

	var frame PopularimeterFrame

	assertFrameSizeMatchesWrite(t, frame)
}

func TestChapterFrameRoundTripV23(t *testing.T) {
	t.Parallel()

	title := TextFrame{Encoding: EncodingISO, Text: "ch"}
	frame := ChapterFrame{
		ElementID:   "chap0",
		StartTime:   time.Duration(1000 * nanosInMillis),
		EndTime:     time.Duration(2000 * nanosInMillis),
		StartOffset: IgnoredOffset,
		EndOffset:   IgnoredOffset,
		Title:       &title,
	}
	assertVersionedFrameSizeMatchesWrite(t, frame, 3)

	got := assertRoundTrip(t, 3, "CHAP", frame).(ChapterFrame)
	if got.ElementID != "chap0" || got.Title == nil || got.Title.Text != "ch" {
		t.Fatalf("got %#v", got)
	}

	if got.Description != nil || got.Link != nil || got.Artwork != nil {
		t.Fatal("absent subframes must stay nil")
	}
}

func TestChapterFrameRoundTripV24(t *testing.T) {
	t.Parallel()

	title := TextFrame{Encoding: EncodingUTF8, Text: "ch"}
	frame := ChapterFrame{
		ElementID: "chap1",
		StartTime: time.Duration(0),
		EndTime:   time.Duration(1000 * nanosInMillis),
		Title:     &title,
	}
	assertVersionedFrameSizeMatchesWrite(t, frame, 4)

	got := assertRoundTrip(t, 4, "CHAP", frame).(ChapterFrame)
	if got.Title == nil || got.Title.Text != "ch" {
		t.Fatalf("title = %#v", got.Title)
	}
}

func TestChapterFrameRoundTripAllSubframes(t *testing.T) {
	t.Parallel()

	title := TextFrame{Encoding: EncodingUTF8, Text: "title"}
	desc := TextFrame{Encoding: EncodingUTF8, Text: "desc"}
	link := LinkFrame{Encoding: EncodingUTF8, URL: "https://example.com"}
	art := PictureFrame{
		Encoding:    EncodingUTF8,
		MimeType:    "image/jpeg",
		PictureType: PTFrontCover,
		Description: "art",
		Picture:     []byte{9, 8, 7},
	}
	frame := ChapterFrame{
		ElementID:   "chap-all",
		StartTime:   time.Duration(10 * nanosInMillis),
		EndTime:     time.Duration(20 * nanosInMillis),
		StartOffset: 1,
		EndOffset:   2,
		Title:       &title,
		Description: &desc,
		Link:        &link,
		Artwork:     &art,
	}
	assertFrameSizeMatchesWrite(t, frame)
	assertVersionedFrameSizeMatchesWrite(t, frame, 4)

	got := assertRoundTrip(t, 4, "CHAP", frame).(ChapterFrame)
	if got.Title == nil || got.Title.Text != "title" {
		t.Fatalf("Title = %#v", got.Title)
	}

	if got.Description == nil || got.Description.Text != "desc" {
		t.Fatalf("Description = %#v", got.Description)
	}

	if got.Link == nil || got.Link.URL != "https://example.com" {
		t.Fatalf("Link = %#v", got.Link)
	}

	if got.Artwork == nil || !bytes.Equal(got.Artwork.Picture, art.Picture) {
		t.Fatalf("Artwork = %#v", got.Artwork)
	}
}

func TestChapterFrameAbsentSubframesRemainNil(t *testing.T) {
	t.Parallel()

	got := assertRoundTrip(t, 4, "CHAP", ChapterFrame{ElementID: "empty"}).(ChapterFrame)
	if got.Title != nil || got.Description != nil || got.Link != nil || got.Artwork != nil {
		t.Fatalf("expected nil subframes, got %#v", got)
	}
}

func TestChapterFrameSkipsUnknownEmbeddedFrame(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	bw := newBufferedWriter(&buf)

	bw.WriteString("chap")
	bw.writeByte(0)

	for range 4 {
		if err := binary.Write(bw, binary.BigEndian, uint32(0)); err != nil {
			t.Fatal(err)
		}
	}

	if err := writeFrameHeader(bw, "PRIV", 4, true); err != nil {
		t.Fatal(err)
	}

	if _, err := bw.Write([]byte("abcd")); err != nil {
		t.Fatal(err)
	}

	title := TextFrame{Encoding: EncodingUTF8, Text: "keep"}
	if err := writeEmbeddedFrame(bw, TitleFrameID, title, 4); err != nil {
		t.Fatal(err)
	}

	if err := bw.Flush(); err != nil {
		t.Fatal(err)
	}

	parsed, err := parseChapterFrame(newBufferedReader(bytes.NewReader(buf.Bytes())), 4)
	if err != nil {
		t.Fatalf("parseChapterFrame: %v", err)
	}

	got := parsed.(ChapterFrame)
	if got.Title == nil || got.Title.Text != "keep" {
		t.Fatalf("Title = %#v after skipping unknown frame", got.Title)
	}
}

func TestChapterFrameTruncatedEmbeddedHeader(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	bw := newBufferedWriter(&buf)

	bw.WriteString("chap")
	bw.writeByte(0)

	for range 4 {
		if err := binary.Write(bw, binary.BigEndian, uint32(0)); err != nil {
			t.Fatal(err)
		}
	}

	if _, err := bw.Write([]byte("TIT2\x00")); err != nil {
		t.Fatal(err)
	}

	if err := bw.Flush(); err != nil {
		t.Fatal(err)
	}

	if _, err := parseChapterFrame(newBufferedReader(bytes.NewReader(buf.Bytes())), 4); err == nil {
		t.Fatal("expected error for truncated embedded header")
	}
}

func TestChapterFrameTruncatedEmbeddedBody(t *testing.T) {
	t.Parallel()

	title := TextFrame{Encoding: EncodingUTF8, Text: "long-title"}
	frame := ChapterFrame{ElementID: "chap", Title: &title}

	var buf bytes.Buffer
	if _, err := frame.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}

	truncated := buf.Bytes()[:buf.Len()-2]
	if _, err := parseChapterFrame(newBufferedReader(bytes.NewReader(truncated)), 4); err == nil {
		t.Fatal("expected error for truncated embedded body")
	}
}

func TestParseHeaderHandlesShortReads(t *testing.T) {
	t.Parallel()

	reader := iotest.OneByteReader(bytes.NewReader(thb))

	header, err := parseHeader(reader)
	if err != nil {
		t.Fatalf("parseHeader: %v", err)
	}

	if header != th {
		t.Fatalf("header = %#v, want %#v", header, th)
	}
}

func TestParseHeaderEmptyReader(t *testing.T) {
	t.Parallel()

	_, err := parseHeader(bytes.NewReader(nil))
	if !errors.Is(err, io.EOF) {
		t.Fatalf("err = %v, want EOF", err)
	}
}

func TestParseHeaderPartialBytes(t *testing.T) {
	t.Parallel()

	for n := 1; n < tagHeaderSize; n++ {
		_, err := parseHeader(bytes.NewReader(thb[:n]))
		if !errors.Is(err, ErrSmallHeaderSize) {
			t.Fatalf("n=%d err = %v, want ErrSmallHeaderSize", n, err)
		}
	}
}

func TestParseFrameHeaderHandlesShortReads(t *testing.T) {
	t.Parallel()

	frame := TextFrame{Encoding: EncodingUTF8, Text: "x"}
	tag := NewEmptyTag()
	tag.AddTextFrame("TIT2", frame.Encoding, frame.Text)

	var buf bytes.Buffer
	if _, err := tag.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}

	rd := iotest.OneByteReader(bytes.NewReader(buf.Bytes()[tagHeaderSize:]))
	headerBuf := make([]byte, frameHeaderSize)

	header, err := parseFrameHeader(headerBuf, rd, true)
	if err != nil {
		t.Fatalf("parseFrameHeader: %v", err)
	}

	if header.ID != "TIT2" {
		t.Fatalf("ID = %q", header.ID)
	}
}

func TestParseReaderTruncatedFrameHeader(t *testing.T) {
	t.Parallel()

	data := append(append([]byte{}, id3Identifier...), 4, 0, 0, 0, 0, 0, 5)
	data = append(data, 'T', 'I', 'T')

	_, err := ParseReader(bytes.NewReader(data), Options{Parse: true})
	if err == nil {
		t.Fatal("expected error for truncated frame header")
	}
}

func TestGetFramesDoesNotExposeInternalSlice(t *testing.T) {
	t.Parallel()

	tag := NewEmptyTag()
	tag.AddCommentFrame(engComm)

	frames := tag.GetFrames(tag.CommonID("Comments"))
	if len(frames) != 1 {
		t.Fatalf("len = %d", len(frames))
	}

	frames[0] = CommentFrame{Language: "xxx", Text: "mutated"}

	got := tag.GetFrames(tag.CommonID("Comments"))[0].(CommentFrame)
	if got.Text != engComm.Text {
		t.Fatalf("internal slice was mutated: %#v", got)
	}
}

func TestMustFrameBeInSequence(t *testing.T) {
	t.Parallel()

	tests := []struct {
		id   string
		want bool
	}{
		{"MCDI", false},
		{"ETCO", false},
		{"SYTC", false},
		{"RVRB", false},
		{"MLLT", false},
		{"PCNT", false},
		{"RBUF", false},
		{"POSS", false},
		{"OWNE", false},
		{"SEEK", false},
		{"ASPI", false},
		{"IPLS", false},
		{"RVAD", false},
		{"APIC", true},
		{"COMM", true},
		{"USLT", true},
		{"SYLT", true},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			if got := mustFrameBeInSequence(tt.id); got != tt.want {
				t.Fatalf("mustFrameBeInSequence(%q) = %v, want %v", tt.id, got, tt.want)
			}
		})
	}
}

func TestSequenceKeepsMultipleUnknownFrames(t *testing.T) {
	t.Parallel()

	s := getSequence()
	defer putSequence(s)

	s.AddFrame(UnknownFrame{Body: []byte("a")})
	s.AddFrame(UnknownFrame{Body: []byte("b")})
	s.AddFrame(&UnknownFrame{Body: []byte("c")})
	s.AddFrame(&UnknownFrame{Body: []byte("d")})

	if s.Count() != 4 {
		t.Fatalf("Count() = %d, want 4", s.Count())
	}
}

func TestOpenClosesFileOnParseError(t *testing.T) {
	path := copyFixture(t)
	if err := os.WriteFile(path, []byte("ID3\x02"), 0o600); err != nil {
		t.Fatal(err)
	}

	tag, err := Open(path, Options{Parse: true})
	if err == nil {
		_ = tag.Close()

		t.Fatal("expected parse error")
	}

	f, openErr := os.OpenFile(path, os.O_RDWR, 0)
	if openErr != nil {
		t.Fatalf("file still locked after failed Open: %v", openErr)
	}

	_ = f.Close()
}
