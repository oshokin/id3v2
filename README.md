# id3v2

Implementation of ID3 v2.3 and v2.4 in native Go.

This is a fork of [https://github.com/n10v/id3v2/v2](https://github.com/n10v/id3v2).  
The original author hasn't updated it in over two years, so I decided to do it myself.

## What's New?
- **Dropped v1 library** – It's outdated, unnecessary, and I don't care about legacy stuff.
- **Removed ignored errors** – No silent failures, better debugging.
- **Updated Go version requirement** – Works with Go **1.23+** because I used this for a modern project. v1 is older than me.
- **Added linters and task files** – Helps with local development and keeps the code clean.
- **Added SYLT frame support** – Enables synchronized lyrics/text in MP3 files.
- **Added LRC file parsing** – Since you're often adding LRC file content to MP3s, this is essential.

I'll try to keep up with updates from the original repository, but no promises — I'm busy as hell.

## Installation

```bash
go get -u github.com/oshokin/id3v2/v2
```

## Documentation

Full documentation is in the `example_test.go` file at the root of the project.

## Usage Example

```go
package main

import (
	"fmt"
	"log"

	"github.com/oshokin/id3v2/v2"
)

func main() {
	// Open file and parse tag.
	tag, err := id3v2.Open("file.mp3", id3v2.Options{Parse: true})
	if err != nil {
		log.Fatal("Error opening MP3 file:", err)
	}
	defer tag.Close() // Ensure the file is closed when done.

	// Read frames.
	fmt.Println(tag.Artist())
	fmt.Println(tag.Title())

	// Set simple text frames.
	tag.SetArtist("Artist")
	tag.SetTitle("Title")

	// Set comment frame.
	comment := id3v2.CommentFrame{
		Encoding:    id3v2.EncodingUTF8,       // Use UTF-8 encoding.
		Language:    id3v2.EnglishISO6392Code, // Language: English.
		Description: "My opinion",             // Brief description.
		Text:        "Very good song",         // The actual comment.
	}

	tag.AddCommentFrame(comment)

	// Write tag to file.
	if err = tag.Save(); err != nil {
		log.Fatal("Error saving tag:", err)
	}
}
```

## Adding a SYLT Frame (Synchronized Lyrics/Text)

```go
// Define the synchronized lyrics in LRC format or load it from a file.
lyrics := `
[00:02.02] Пусть проходит туман
[00:05.88] Моих призрачных дней
[00:11.56] Пусть заполнит меня дурман
[00:18.30] Лишь бы не думать о ней
[00:22.69] Ведь тебя рядом нет
[00:39.28] И мне не по себе
[00:45.40] Ведь тебя не найти мне
[00:52.85] В безликой толпе
[00:55.49]
`

// Parse the LRC file content into a SynchronisedLyricsFrame.
result, err := id3v2.ParseLRCFile(strings.NewReader(lyrics))
if err != nil {
	log.Fatal("Error parsing LRC file:", err)
}

// Create a SynchronisedLyricsFrame from the parsed result.
sylf := id3v2.SynchronisedLyricsFrame{
	Encoding:          id3v2.EncodingUTF8,
	Language:          id3v2.RussianISO6392Code,                      // Language: Russian.
	TimestampFormat:   id3v2.SYLTAbsoluteMillisecondsTimestampFormat, // Use absolute timestamps.
	ContentType:       id3v2.SYLTLyricsContentType,                   // Mark as lyrics.
	ContentDescriptor: "Lyrics",                                      // Descriptor for lyrics.
	SynchronizedTexts: result.SynchronizedTexts,                      // The actual synchronized lyrics.
}

// Add the synchronized lyrics frame to the tag.
tag.AddSynchronisedLyricsFrame(sylf)
```
