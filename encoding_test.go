package id3v2

import (
	"bytes"
	"sync"
	"testing"
)

func TestDecodeText(t *testing.T) {
	testCases := []struct {
		src  []byte
		from Encoding
		utf8 string
	}{
		{[]byte{0x48, 0xE9, 0x6C, 0x6C, 0xF6}, EncodingISO, "Héllö"},
		{[]byte{0xFF, 0xFE, 0x48, 0x00, 0xE9, 0x00, 0x6C, 0x00, 0x6C, 0x00, 0xF6, 0x00}, EncodingUTF16, "Héllö"},
		{[]byte{0xFF, 0xFE}, EncodingUTF16, ""},
		{[]byte{0x00, 0x48, 0x00, 0xE9, 0x00, 0x6C, 0x00, 0x6C, 0x00, 0xF6}, EncodingUTF16BE, "Héllö"},
	}

	for _, tc := range testCases {
		got := decodeText(tc.src, tc.from)
		if got != tc.utf8 {
			t.Errorf("Expected %q from %v encoding, got %q", tc.utf8, tc.from, got)
		}
	}
}

func TestDecodeTextParallel(t *testing.T) {
	testCases := []struct {
		src  []byte
		from Encoding
		utf8 string
	}{
		{[]byte{0x48, 0xE9, 0x6C, 0x6C, 0xF6}, EncodingISO, "Héllö"},
		{[]byte{0x48, 0xE9, 0x6C, 0x6C, 0xF6}, EncodingISO, "Héllö"},
		{[]byte{0x48, 0xE9, 0x6C, 0x6C, 0xF6}, EncodingISO, "Héllö"},
		{[]byte{0x48, 0xE9, 0x6C, 0x6C, 0xF6}, EncodingISO, "Héllö"},
		{[]byte{0x48, 0xE9, 0x6C, 0x6C, 0xF6}, EncodingISO, "Héllö"},
		{[]byte{0xFF, 0xFE, 0x48, 0x00, 0xE9, 0x00, 0x6C, 0x00, 0x6C, 0x00, 0xF6, 0x00}, EncodingUTF16, "Héllö"},
		{[]byte{0xFF, 0xFE, 0x48, 0x00, 0xE9, 0x00, 0x6C, 0x00, 0x6C, 0x00, 0xF6, 0x00}, EncodingUTF16, "Héllö"},
		{[]byte{0xFF, 0xFE, 0x48, 0x00, 0xE9, 0x00, 0x6C, 0x00, 0x6C, 0x00, 0xF6, 0x00}, EncodingUTF16, "Héllö"},
		{[]byte{0xFF, 0xFE, 0x48, 0x00, 0xE9, 0x00, 0x6C, 0x00, 0x6C, 0x00, 0xF6, 0x00}, EncodingUTF16, "Héllö"},
		{[]byte{0xFF, 0xFE, 0x48, 0x00, 0xE9, 0x00, 0x6C, 0x00, 0x6C, 0x00, 0xF6, 0x00}, EncodingUTF16, "Héllö"},
		{[]byte{0xFF, 0xFE}, EncodingUTF16, ""},
		{[]byte{0xFF, 0xFE}, EncodingUTF16, ""},
		{[]byte{0xFF, 0xFE}, EncodingUTF16, ""},
		{[]byte{0xFF, 0xFE}, EncodingUTF16, ""},
		{[]byte{0xFF, 0xFE}, EncodingUTF16, ""},
		{[]byte{0x00, 0x48, 0x00, 0xE9, 0x00, 0x6C, 0x00, 0x6C, 0x00, 0xF6}, EncodingUTF16BE, "Héllö"},
		{[]byte{0x00, 0x48, 0x00, 0xE9, 0x00, 0x6C, 0x00, 0x6C, 0x00, 0xF6}, EncodingUTF16BE, "Héllö"},
		{[]byte{0x00, 0x48, 0x00, 0xE9, 0x00, 0x6C, 0x00, 0x6C, 0x00, 0xF6}, EncodingUTF16BE, "Héllö"},
		{[]byte{0x00, 0x48, 0x00, 0xE9, 0x00, 0x6C, 0x00, 0x6C, 0x00, 0xF6}, EncodingUTF16BE, "Héllö"},
		{[]byte{0x00, 0x48, 0x00, 0xE9, 0x00, 0x6C, 0x00, 0x6C, 0x00, 0xF6}, EncodingUTF16BE, "Héllö"},
	}

	var wg sync.WaitGroup

	for _, tc := range testCases {
		wg.Go(func() {
			got := decodeText(tc.src, tc.from)
			if got != tc.utf8 {
				t.Errorf("Expected %q from %v encoding, got %q", tc.utf8, tc.from, got)
			}
		})
	}

	wg.Wait()
}

func TestDecodeMultiUTF16BEAlignedTerminator(t *testing.T) {
	t.Parallel()

	// Ā = 01 00, terminator = 00 00, B = 00 42, terminator = 00 00.
	src := []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x42, 0x00, 0x00}

	got := decodeMulti(src, EncodingUTF16BE)
	if len(got) != 2 || got[0] != "Ā" || got[1] != "B" {
		t.Fatalf("decodeMulti = %#v, want [Ā B]", got)
	}
}

func TestParseTextFrameUTF16LEBOMMultiValue(t *testing.T) {
	t.Parallel()

	// Encoding UTF-16 with a little-endian BOM, then Ā and B.
	payload := []byte{
		1,
		0xFF, 0xFE,
		0x00, 0x01,
		0x00, 0x00,
		0x42, 0x00,
		0x00, 0x00,
	}

	got, err := parseTextFrame(newBufferedReader(bytes.NewReader(payload)))
	if err != nil {
		t.Fatal(err)
	}

	tf, ok := got.(TextFrame)
	if !ok {
		t.Fatalf("got %T", got)
	}

	if tf.Text != "Ā" || len(tf.Multi) != 2 || tf.Multi[0] != "Ā" || tf.Multi[1] != "B" {
		t.Fatalf("got Text=%q Multi=%#v", tf.Text, tf.Multi)
	}
}

func TestUTF16ReplacementCharacterPreserved(t *testing.T) {
	const text = "abc\uFFFDdef"

	// UTF-16LE with BOM: "abc" + U+FFFD + "def".
	src := []byte{
		0xFF, 0xFE,
		0x61, 0x00, 0x62, 0x00, 0x63, 0x00,
		0xFD, 0xFF,
		0x64, 0x00, 0x65, 0x00, 0x66, 0x00,
	}

	got := decodeText(src, EncodingUTF16)
	if got != text {
		t.Fatalf("decodeText: got %q, want %q", got, text)
	}

	buf := new(bytes.Buffer)
	bw := newBufferedWriter(buf)

	if err := encodeWriteText(bw, text, EncodingUTF16); err != nil {
		t.Fatal(err)
	}

	if err := bw.Flush(); err != nil {
		t.Fatal(err)
	}

	got = decodeText(buf.Bytes(), EncodingUTF16)
	if got != text {
		t.Fatalf("round-trip: got %q, want %q", got, text)
	}
}

func TestEncodeWriteText(t *testing.T) {
	testCases := []struct {
		src      string
		to       Encoding
		expected []byte
	}{
		{"Héllö", EncodingISO, []byte{0x48, 0xE9, 0x6C, 0x6C, 0xF6}},
		{"Héllö", EncodingUTF16, []byte{0xFE, 0xFF, 0x00, 0x48, 0x00, 0xE9, 0x00, 0x6C, 0x00, 0x6C, 0x00, 0xF6}},
		{"Héllö", EncodingUTF16BE, []byte{0x00, 0x48, 0x00, 0xE9, 0x00, 0x6C, 0x00, 0x6C, 0x00, 0xF6}},
	}

	buf := new(bytes.Buffer)
	bw := newBufferedWriter(buf)

	for _, tc := range testCases {
		buf.Reset()
		bw.Reset(buf)

		bw.EncodeAndWriteText(tc.src, tc.to)

		if err := bw.Flush(); err != nil {
			t.Fatal(err)
		}

		got := buf.Bytes()
		if !bytes.Equal(got, tc.expected) {
			t.Errorf("Expected %q to %q encoding, got %q", tc.expected, tc.to, got)
		}

		if bw.Written() != len(tc.expected) {
			t.Errorf("Expected %v size, got %v", len(tc.expected), bw.Written())
		}
	}
}

func TestUnsynchronisedLyricsFrameWithUTF16(t *testing.T) {
	contentDescriptor := "Content descriptor"
	lyrics := "Lyrics"

	frame := UnsynchronisedLyricsFrame{
		Encoding:          EncodingUTF16,
		Language:          EnglishISO6392Code,
		ContentDescriptor: contentDescriptor,
		Lyrics:            lyrics,
	}

	buf := new(bytes.Buffer)

	if _, err := frame.WriteTo(buf); err != nil {
		t.Fatal(err)
	}

	parsed, err := parseUnsynchronisedLyricsFrame(newBufferedReader(buf), 4)
	if err != nil {
		t.Fatal(err)
	}

	uslf := parsed.(UnsynchronisedLyricsFrame)

	if uslf.ContentDescriptor != contentDescriptor {
		t.Errorf("Expected content descriptor: %q, got: %q", contentDescriptor, uslf.ContentDescriptor)
	}

	if uslf.Lyrics != lyrics {
		t.Errorf("Expected lyrics: %q, got: %q", lyrics, uslf.Lyrics)
	}
}
