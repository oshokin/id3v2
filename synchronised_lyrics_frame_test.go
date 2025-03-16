package id3v2

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestParseLRCFile(t *testing.T) {
	lrcContent := `
[ar:Artist Name]
[al:Album Name]
[ti:Title]
[au:Funny Guy]
[lr:Joke Master]
[length:03:30]
[by:Lyrics Guru]
[tool:Go LRC Parser]
[ve:1.0]
[offset:500]
[00:10.00]I’m a banana, I’m a banana, I’m a banana, look at me move!
[00:20.00]I’m a cucumber, I’m a cucumber, I’m a cucumber, I’m in a groove!
[00:30.00]I’m a tomato, I’m a tomato, I’m a tomato, I’m feeling fruity!
[00:40.00]I’m a potato, I’m a potato, I’m a potato, I’m kinda booty!
`

	reader := strings.NewReader(lrcContent)

	result, err := ParseLRCFile(reader)
	if err != nil {
		t.Fatalf("Error parsing LRC file: %v", err)
	}

	if result.Metadata[LRCTagArtist] != "Artist Name" {
		t.Errorf("Expected artist metadata 'Artist Name', got '%s'", result.Metadata["ar"])
	}
	if result.Metadata[LRCTagAlbum] != "Album Name" {
		t.Errorf("Expected album metadata 'Album Name', got '%s'", result.Metadata["al"])
	}
	if result.Metadata[LRCTagTitle] != "Title" {
		t.Errorf("Expected title metadata 'Title', got '%s'", result.Metadata["ti"])
	}
	if result.Metadata[LRCTagAuthor] != "Funny Guy" {
		t.Errorf("Expected author metadata 'Funny Guy', got '%s'", result.Metadata["au"])
	}
	if result.Metadata[LRCTagLyricist] != "Joke Master" {
		t.Errorf("Expected lyricist metadata 'Joke Master', got '%s'", result.Metadata["lr"])
	}
	if result.Metadata[LRCTagLength] != "03:30" {
		t.Errorf("Expected length metadata '03:30', got '%s'", result.Metadata["length"])
	}
	if result.Metadata[LRCTagBy] != "Lyrics Guru" {
		t.Errorf("Expected LRC author metadata 'Lyrics Guru', got '%s'", result.Metadata["by"])
	}
	if result.Metadata[LRCTagTool] != "Go LRC Parser" {
		t.Errorf("Expected tool metadata 'Go LRC Parser', got '%s'", result.Metadata["tool"])
	}
	if result.Metadata[LRCTagVersion] != "1.0" {
		t.Errorf("Expected version metadata '1.0', got '%s'", result.Metadata["ve"])
	}

	if result.TimestampFormat != SYLTAbsoluteMillisecondsTimestampFormat {
		t.Errorf("Expected timestamp format SYLTAbsoluteMillisecondsTimestampFormat, got %v", result.TimestampFormat)
	}

	expectedLyrics := []SynchronizedText{
		{Text: "I’m a banana, I’m a banana, I’m a banana, look at me move!", Timestamp: 10500},
		{Text: "I’m a cucumber, I’m a cucumber, I’m a cucumber, I’m in a groove!", Timestamp: 20500},
		{Text: "I’m a tomato, I’m a tomato, I’m a tomato, I’m feeling fruity!", Timestamp: 30500},
		{Text: "I’m a potato, I’m a potato, I’m a potato, I’m kinda booty!", Timestamp: 40500},
	}

	if len(result.SynchronizedTexts) != len(expectedLyrics) {
		t.Fatalf("Expected %d synchronized lyrics, got %d", len(expectedLyrics), len(result.SynchronizedTexts))
	}

	for i, expected := range expectedLyrics {
		if result.SynchronizedTexts[i].Text != expected.Text {
			t.Errorf("Expected lyric text '%s', got '%s'", expected.Text, result.SynchronizedTexts[i].Text)
		}

		if result.SynchronizedTexts[i].Timestamp != expected.Timestamp {
			t.Errorf("Expected timestamp %d, got %d", expected.Timestamp, result.SynchronizedTexts[i].Timestamp)
		}
	}
}

func TestSynchronisedLyricsFrameWriteTo(t *testing.T) {
	sylf := SynchronisedLyricsFrame{
		Encoding:          EncodingUTF8,
		Language:          EnglishISO6392Code,
		TimestampFormat:   SYLTAbsoluteMillisecondsTimestampFormat,
		ContentType:       SYLTLyricsContentType,
		ContentDescriptor: "Verse 1",
		SynchronizedTexts: []SynchronizedText{
			{Text: "First line", Timestamp: 1000},
			{Text: "Second line", Timestamp: 2000},
		},
	}

	buf := new(bytes.Buffer)
	_, err := sylf.WriteTo(buf)
	if err != nil {
		t.Fatalf("Error writing SynchronisedLyricsFrame: %v", err)
	}

	br := newBufferedReader(buf)
	parsedFrame, err := parseSynchronisedLyricsFrame(br, 4)
	if err != nil {
		t.Fatalf("Error parsing SynchronisedLyricsFrame: %v", err)
	}

	parsedSylf, ok := parsedFrame.(SynchronisedLyricsFrame)
	if !ok {
		t.Fatal("Parsed frame is not a SynchronisedLyricsFrame")
	}

	if !parsedSylf.Encoding.Equals(sylf.Encoding) {
		t.Errorf("Expected encoding %v, got %v", sylf.Encoding, parsedSylf.Encoding)
	}

	if parsedSylf.Language != sylf.Language {
		t.Errorf("Expected language %v, got %v", sylf.Language, parsedSylf.Language)
	}

	if parsedSylf.TimestampFormat != sylf.TimestampFormat {
		t.Errorf("Expected timestamp format %v, got %v", sylf.TimestampFormat, parsedSylf.TimestampFormat)
	}

	if parsedSylf.ContentType != sylf.ContentType {
		t.Errorf("Expected content type %v, got %v", sylf.ContentType, parsedSylf.ContentType)
	}

	if parsedSylf.ContentDescriptor != sylf.ContentDescriptor {
		t.Errorf("Expected content descriptor %v, got %v", sylf.ContentDescriptor, parsedSylf.ContentDescriptor)
	}

	if len(parsedSylf.SynchronizedTexts) != len(sylf.SynchronizedTexts) {
		t.Fatalf("Expected %d synchronized texts, got %d", len(sylf.SynchronizedTexts), len(parsedSylf.SynchronizedTexts))
	}

	for i, expected := range sylf.SynchronizedTexts {
		if parsedSylf.SynchronizedTexts[i].Text != expected.Text {
			t.Errorf("Expected text '%s', got '%s'", expected.Text, parsedSylf.SynchronizedTexts[i].Text)
		}

		if parsedSylf.SynchronizedTexts[i].Timestamp != expected.Timestamp {
			t.Errorf("Expected timestamp %d, got %d", expected.Timestamp, parsedSylf.SynchronizedTexts[i].Timestamp)
		}
	}
}

func TestSynchronisedLyricsFrameSize(t *testing.T) {
	sylf := SynchronisedLyricsFrame{
		Encoding:          EncodingUTF8,
		Language:          EnglishISO6392Code,
		TimestampFormat:   SYLTAbsoluteMillisecondsTimestampFormat,
		ContentType:       SYLTLyricsContentType,
		ContentDescriptor: "Verse 1",
		SynchronizedTexts: []SynchronizedText{
			{Text: "First line", Timestamp: 1000},
			{Text: "Second line", Timestamp: 2000},
		},
	}

	expectedSize := 1 + // Encoding byte
		len(sylf.Language) + // Language code (3 bytes)
		encodedSize(sylf.ContentDescriptor, sylf.Encoding) + // Content descriptor
		len(sylf.Encoding.TerminationBytes) + // Termination bytes for descriptor
		1 + // Timestamp format byte
		1 + // Content type byte
		encodedSize("First line", sylf.Encoding) + // First line text
		len(sylf.Encoding.TerminationBytes) + // Termination bytes for first line
		4 + // First line timestamp (4 bytes)
		encodedSize("Second line", sylf.Encoding) + // Second line text
		len(sylf.Encoding.TerminationBytes) + // Termination bytes for second line
		4 // Second line timestamp (4 bytes)

	if sylf.Size() != expectedSize {
		t.Errorf("Expected size %d, got %d", expectedSize, sylf.Size())
	}
}

func TestAddSynchronisedLyricsFrame(t *testing.T) {
	type fields struct {
		Encoding          Encoding
		Language          string
		TimestampFormat   SYLTTimestampFormat
		ContentType       SYLTContentType
		ContentDescriptor string
		SynchronizedTexts []SynchronizedText
	}

	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "basic SYLT frame",
			fields: fields{
				Encoding:          EncodingUTF8,
				Language:          EnglishISO6392Code,
				TimestampFormat:   SYLTAbsoluteMillisecondsTimestampFormat,
				ContentType:       SYLTLyricsContentType,
				ContentDescriptor: "Verse 1",
				SynchronizedTexts: []SynchronizedText{
					{Text: "First line", Timestamp: 1000},
					{Text: "Second line", Timestamp: 2000},
				},
			},
		},
		{
			name: "SYLT frame with empty lyrics",
			fields: fields{
				Encoding:          EncodingUTF8,
				Language:          EnglishISO6392Code,
				TimestampFormat:   SYLTAbsoluteMillisecondsTimestampFormat,
				ContentType:       SYLTLyricsContentType,
				ContentDescriptor: "Empty Verse",
				SynchronizedTexts: []SynchronizedText{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile, err := prepareTestFile("sylt_test")
			if err != nil {
				t.Error(err)
			}
			defer os.Remove(tmpFile.Name())

			tag, err := Open(tmpFile.Name(), Options{Parse: true})
			if tag == nil || err != nil {
				t.Fatal("Error while opening mp3 file: ", err)
			}

			sylf := SynchronisedLyricsFrame{
				Encoding:          tt.fields.Encoding,
				Language:          tt.fields.Language,
				TimestampFormat:   tt.fields.TimestampFormat,
				ContentType:       tt.fields.ContentType,
				ContentDescriptor: tt.fields.ContentDescriptor,
				SynchronizedTexts: tt.fields.SynchronizedTexts,
			}
			tag.AddSynchronisedLyricsFrame(sylf)

			if err = tag.Save(); err != nil {
				t.Error(err)
			}

			tag.Close()

			tag, err = Open(tmpFile.Name(), Options{Parse: true})
			if tag == nil || err != nil {
				t.Fatal("Error while opening mp3 file: ", err)
			}
			defer tag.Close()

			frame := tag.GetLastFrame(tag.CommonID("Synchronised lyrics/text"))
			if frame == nil {
				t.Fatal("SYLT frame not found in the tag")
			}

			parsedSylf, ok := frame.(SynchronisedLyricsFrame)
			if !ok {
				t.Fatal("Could not assert frame as SynchronisedLyricsFrame")
			}

			if !parsedSylf.Encoding.Equals(tt.fields.Encoding) {
				t.Errorf("Expected encoding %v, got %v", tt.fields.Encoding, parsedSylf.Encoding)
			}

			if parsedSylf.Language != tt.fields.Language {
				t.Errorf("Expected language %v, got %v", tt.fields.Language, parsedSylf.Language)
			}

			if parsedSylf.TimestampFormat != tt.fields.TimestampFormat {
				t.Errorf("Expected timestamp format %v, got %v", tt.fields.TimestampFormat, parsedSylf.TimestampFormat)
			}

			if parsedSylf.ContentType != tt.fields.ContentType {
				t.Errorf("Expected content type %v, got %v", tt.fields.ContentType, parsedSylf.ContentType)
			}

			if parsedSylf.ContentDescriptor != tt.fields.ContentDescriptor {
				t.Errorf("Expected content descriptor %v, got %v", tt.fields.ContentDescriptor, parsedSylf.ContentDescriptor)
			}

			if len(parsedSylf.SynchronizedTexts) != len(tt.fields.SynchronizedTexts) {
				t.Errorf("Expected %d synchronized texts, got %d", len(tt.fields.SynchronizedTexts), len(parsedSylf.SynchronizedTexts))
			}

			for i, expectedText := range tt.fields.SynchronizedTexts {
				if parsedSylf.SynchronizedTexts[i].Text != expectedText.Text {
					t.Errorf("Expected text %v, got %v", expectedText.Text, parsedSylf.SynchronizedTexts[i].Text)
				}

				if parsedSylf.SynchronizedTexts[i].Timestamp != expectedText.Timestamp {
					t.Errorf("Expected timestamp %v, got %v", expectedText.Timestamp, parsedSylf.SynchronizedTexts[i].Timestamp)
				}
			}
		})
	}
}
