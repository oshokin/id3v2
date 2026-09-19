package id3v2

import (
	"bytes"
	"strings"
	"testing"
)

func FuzzParseReaderDoesNotPanic(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte("ID3"))
	f.Add(thb)
	f.Add([]byte{'I', 'D', '3', 4, 0, 0, 0, 0, 0, 10, 'T'})
	f.Add([]byte{0, 1, 2, 3, 4})
	f.Add(mustReadFile(mp3Path)[:tagHeaderSize])

	f.Fuzz(func(t *testing.T, data []byte) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("parser panicked for %x: %v", data, r)
			}
		}()

		tag, _ := ParseReader(bytes.NewReader(data), Options{Parse: true})
		if tag != nil {
			_ = tag.Close()
		}
	})
}

func FuzzParseSynchronisedLyricsFrameDoesNotPanic(f *testing.F) {
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
		f.Fatal(err)
	}

	f.Add(buf.Bytes())
	f.Add([]byte{})
	f.Add([]byte{3, 'e', 'n', 'g'})

	f.Fuzz(func(t *testing.T, data []byte) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("SYLT parser panicked for %x: %v", data, r)
			}
		}()

		_, _ = parseSynchronisedLyricsFrame(newBufferedReader(bytes.NewReader(data)), 4)
	})
}

func FuzzParseLRCFileDoesNotPanic(f *testing.F) {
	f.Add("")
	f.Add("[00:10.00]hello")
	f.Add("[offset:500]\n[00:01.00]x")

	f.Fuzz(func(t *testing.T, data string) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("LRC parser panicked: %v", r)
			}
		}()

		_, _ = ParseLRCFile(strings.NewReader(data))
	})
}
