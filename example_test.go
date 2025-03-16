package id3v2_test

import (
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"sync"

	id3v2 "github.com/oshokin/id3v2/v2"
)

func Example() {
	// Open file and parse tag in it.
	tag, err := id3v2.Open("file.mp3", id3v2.Options{Parse: true})
	if err != nil {
		log.Fatal("Error while opening mp3 file: ", err)
	}
	defer tag.Close() // Ensure the file is closed when we're done.

	// Read frames.
	fmt.Println(tag.Artist())
	fmt.Println(tag.Title())

	// Set simple text frames.
	tag.SetArtist("Artist")
	tag.SetTitle("Title")

	// Set comment frame.
	comment := id3v2.CommentFrame{
		Encoding:    id3v2.EncodingUTF8,       // Use UTF-8 encoding for the text.
		Language:    id3v2.EnglishISO6392Code, // Specify the language as English.
		Description: "My opinion",             // A brief description of the comment.
		Text:        "Very good song",         // The actual comment text.
	}
	tag.AddCommentFrame(comment)

	// Write tag to file.
	if err = tag.Save(); err != nil {
		log.Fatal("Error while saving a tag: ", err)
	}
}

func Example_concurrent() {
	// Create a pool of reusable tag objects to avoid allocating new ones repeatedly.
	tagPool := sync.Pool{New: func() any {
		return id3v2.NewEmptyTag()
	}}

	var wg sync.WaitGroup // Used to wait for all goroutines to finish.

	wg.Add(100) // We'll spawn 100 goroutines.

	for range 100 {
		go func() {
			defer wg.Done() // Signal that this goroutine is done.

			// Get a tag from the pool (or create a new one if the pool is empty).
			tag := tagPool.Get().(*id3v2.Tag)
			defer tagPool.Put(tag) // Return the tag to the pool when done.

			file, err := os.Open("file.mp3")
			if err != nil {
				log.Fatal("Error while opening file:", err)
			}
			defer file.Close() // Ensure the file is closed when done.

			// Reset the tag to read from the new file.
			if err := tag.Reset(file, id3v2.Options{Parse: true}); err != nil {
				log.Fatal("Error while reseting tag to file:", err)
			}

			// Print the artist and title from the tag.
			fmt.Println(tag.Artist() + " - " + tag.Title())
		}()
	}

	wg.Wait() // Wait for all goroutines to finish.
}

func ExampleTag_AddAttachedPicture() {
	tag, err := id3v2.Open("file.mp3", id3v2.Options{Parse: true})
	if tag == nil || err != nil {
		log.Fatal("Error while opening mp3 file: ", err)
	}

	// Read the artwork file into a byte slice.
	artwork, err := os.ReadFile("artwork.jpg")
	if err != nil {
		log.Fatal("Error while reading artwork file", err)
	}

	// Create a picture frame with the artwork.
	pic := id3v2.PictureFrame{
		Encoding:    id3v2.EncodingUTF8, // Use UTF-8 encoding for the description.
		MimeType:    "image/jpeg",       // Specify the MIME type of the image.
		PictureType: id3v2.PTFrontCover, // Indicate that this is the front cover image.
		Description: "Front cover",      // A description of the picture.
		Picture:     artwork,            // The actual image data.
	}
	tag.AddAttachedPicture(pic)
}

func ExampleTag_AddCommentFrame() {
	tag, err := id3v2.Open("file.mp3", id3v2.Options{Parse: true})
	if tag == nil || err != nil {
		log.Fatal("Error while opening mp3 file: ", err)
	}

	// Create a comment frame with a description and text.
	comment := id3v2.CommentFrame{
		Encoding:    id3v2.EncodingUTF8,
		Language:    id3v2.EnglishISO6392Code,
		Description: "My opinion",
		Text:        "Very good song",
	}
	tag.AddCommentFrame(comment)
}

func ExampleTag_AddUnsynchronisedLyricsFrame() {
	tag, err := id3v2.Open("file.mp3", id3v2.Options{Parse: true})
	if tag == nil || err != nil {
		log.Fatal("Error while opening mp3 file: ", err)
	}

	// Create a frame for unsynchronized lyrics (e.g., lyrics without timestamps).
	uslt := id3v2.UnsynchronisedLyricsFrame{
		Encoding:          id3v2.EncodingUTF8,
		Language:          id3v2.GermanISO6392Code,               // Specify the language as German.
		ContentDescriptor: "Deutsche Nationalhymne",              // A descriptor for the lyrics.
		Lyrics:            "Einigkeit und Recht und Freiheit...", // The actual lyrics.
	}
	tag.AddUnsynchronisedLyricsFrame(uslt)
}

func ExampleTag_GetFrames() {
	tag, err := id3v2.Open("file.mp3", id3v2.Options{Parse: true})
	if tag == nil || err != nil {
		log.Fatal("Error while opening mp3 file: ", err)
	}

	// Retrieve all frames of type "Attached picture".
	pictures := tag.GetFrames(tag.CommonID("Attached picture"))
	for _, f := range pictures {
		// Assert that the frame is a PictureFrame.
		pic, ok := f.(id3v2.PictureFrame)
		if !ok {
			log.Fatal("Couldn't assert picture frame")
		}
		// Do something with the picture frame, e.g., print its description.
		fmt.Println(pic.Description)
	}
}

func ExampleTag_GetLastFrame() {
	tag, err := id3v2.Open("file.mp3", id3v2.Options{Parse: true})
	if tag == nil || err != nil {
		log.Fatal("Error while opening mp3 file: ", err)
	}

	// Retrieve the last frame of type "BPM" (beats per minute).
	bpmFramer := tag.GetLastFrame(tag.CommonID("BPM"))
	if bpmFramer != nil {
		// Assert that the frame is a TextFrame.
		bpm, ok := bpmFramer.(id3v2.TextFrame)
		if !ok {
			log.Fatal("Couldn't assert bpm frame")
		}

		// Print the BPM value.
		fmt.Println(bpm.Text)
	}
}

func ExampleCommentFrame_get() {
	tag, err := id3v2.Open("file.mp3", id3v2.Options{Parse: true})
	if tag == nil || err != nil {
		log.Fatal("Error while opening mp3 file: ", err)
	}

	// Retrieve all comment frames.
	comments := tag.GetFrames(tag.CommonID("Comments"))
	for _, f := range comments {
		// Assert that the frame is a CommentFrame.
		comment, ok := f.(id3v2.CommentFrame)
		if !ok {
			log.Fatal("Couldn't assert comment frame")
		}

		// Do something with the comment, e.g., print its text.
		fmt.Println(comment.Text)
	}
}

func ExampleCommentFrame_add() {
	// Create a new empty tag.
	tag := id3v2.NewEmptyTag()
	// Add a comment frame to the tag.
	comment := id3v2.CommentFrame{
		Encoding:    id3v2.EncodingUTF8,
		Language:    id3v2.EnglishISO6392Code,
		Description: "My opinion",
		Text:        "Very good song",
	}
	tag.AddCommentFrame(comment)
}

func ExamplePictureFrame_get() {
	tag, err := id3v2.Open("file.mp3", id3v2.Options{Parse: true})
	if tag == nil || err != nil {
		log.Fatal("Error while opening mp3 file: ", err)
	}

	// Retrieve all picture frames.
	pictures := tag.GetFrames(tag.CommonID("Attached picture"))
	for _, f := range pictures {
		// Assert that the frame is a PictureFrame.
		pic, ok := f.(id3v2.PictureFrame)
		if !ok {
			log.Fatal("Couldn't assert picture frame")
		}

		// Do something with the picture frame, e.g., print its description.
		fmt.Println(pic.Description)
	}
}

func ExamplePictureFrame_add() {
	// Create a new empty tag.
	tag := id3v2.NewEmptyTag()

	// Read the artwork file into a byte slice.
	artwork, err := os.ReadFile("artwork.jpg")
	if err != nil {
		log.Fatal("Error while reading artwork file", err)
	}

	// Add a picture frame to the tag.
	pic := id3v2.PictureFrame{
		Encoding:    id3v2.EncodingUTF8,
		MimeType:    "image/jpeg",
		PictureType: id3v2.PTFrontCover,
		Description: "Front cover",
		Picture:     artwork,
	}
	tag.AddAttachedPicture(pic)
}

func ExampleTextFrame_get() {
	tag, err := id3v2.Open("file.mp3", id3v2.Options{Parse: true})
	if tag == nil || err != nil {
		log.Fatal("Error while opening mp3 file: ", err)
	}

	// Retrieve a text frame for the "Mood" field.
	tf := tag.GetTextFrame(tag.CommonID("Mood"))
	fmt.Println(tf.Text)
}

func ExampleTextFrame_add() {
	// Create a new empty tag.
	tag := id3v2.NewEmptyTag()
	// Add a text frame for the "Mood" field.
	textFrame := id3v2.TextFrame{
		Encoding: id3v2.EncodingUTF8,
		Text:     "Happy",
	}
	tag.AddFrame(tag.CommonID("Mood"), textFrame)
}

func ExampleUnsynchronisedLyricsFrame_get() {
	tag, err := id3v2.Open("file.mp3", id3v2.Options{Parse: true})
	if tag == nil || err != nil {
		log.Fatal("Error while opening mp3 file: ", err)
	}

	// Retrieve all unsynchronized lyrics frames.
	uslfs := tag.GetFrames(tag.CommonID("Unsynchronised lyrics/text transcription"))
	for _, f := range uslfs {
		// Assert that the frame is an UnsynchronisedLyricsFrame.
		uslf, ok := f.(id3v2.UnsynchronisedLyricsFrame)
		if !ok {
			log.Fatal("Couldn't assert USLT frame")
		}

		// Do something with the lyrics, e.g., print them.
		fmt.Println(uslf.Lyrics)
	}
}

func ExampleUnsynchronisedLyricsFrame_add() {
	// Create a new empty tag.
	tag := id3v2.NewEmptyTag()
	// Add an unsynchronized lyrics frame to the tag.
	uslt := id3v2.UnsynchronisedLyricsFrame{
		Encoding:          id3v2.EncodingUTF8,
		Language:          id3v2.GermanISO6392Code,
		ContentDescriptor: "Deutsche Nationalhymne",
		Lyrics:            "Einigkeit und Recht und Freiheit...",
	}
	tag.AddUnsynchronisedLyricsFrame(uslt)
}

func ExamplePopularimeterFrame_add() {
	// Create a new empty tag.
	tag := id3v2.NewEmptyTag()

	// Add a Popularimeter frame (used for ratings and play counts).
	popmFrame := id3v2.PopularimeterFrame{
		Email:   "foo@bar.com",                 // The email associated with the rating.
		Rating:  128,                           // The rating value (0-255).
		Counter: big.NewInt(10000000000000000), // The play count.
	}
	tag.AddFrame(tag.CommonID("Popularimeter"), popmFrame)
}

func ExamplePopularimeterFrame_get() {
	tag, err := id3v2.Open("file.mp3", id3v2.Options{Parse: true})
	if tag == nil || err != nil {
		log.Fatal("Error while opening mp3 file: ", err)
	}

	// Retrieve the last Popularimeter frame.
	f := tag.GetLastFrame(tag.CommonID("Popularimeter"))

	// Assert that the frame is a PopularimeterFrame.
	popm, ok := f.(id3v2.PopularimeterFrame)
	if !ok {
		log.Fatal("Couldn't assert POPM frame")
	}

	// Do something with the Popularimeter frame, e.g., print its details.
	fmt.Printf("Email: %s, Rating: %d, Counter: %d", popm.Email, popm.Rating, popm.Counter)
}

func ExampleSynchronisedLyricsFrame_add() {
	// Create a new empty tag.
	tag := id3v2.NewEmptyTag()

	// Define the synchronized lyrics in LRC format.
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
		Language:          id3v2.RussianISO6392Code,                      // Specify the language as Russian.
		TimestampFormat:   id3v2.SYLTAbsoluteMillisecondsTimestampFormat, // Use absolute timestamps.
		ContentType:       id3v2.SYLTLyricsContentType,                   // Indicate that this is a lyrics frame.
		ContentDescriptor: "Lyrics",                                      // A descriptor for the lyrics.
		SynchronizedTexts: result.SynchronizedTexts,                      // The actual synchronized lyrics.
	}

	// Add the synchronized lyrics frame to the tag.
	tag.AddSynchronisedLyricsFrame(sylf)
}

func ExampleSynchronisedLyricsFrame_get() {
	// Open the file and parse the tag.
	tag, err := id3v2.Open("file.mp3", id3v2.Options{Parse: true})
	if tag == nil || err != nil {
		log.Fatal("Error opening MP3 file:", err)
	}
	defer tag.Close()

	// Retrieve the last synchronized lyrics frame.
	frame := tag.GetLastFrame(tag.CommonID("Synchronised lyrics/text"))
	if frame == nil {
		log.Fatal("No synchronized lyrics frame found")
	}

	// Assert that the frame is a SynchronisedLyricsFrame.
	sylf, ok := frame.(id3v2.SynchronisedLyricsFrame)
	if !ok {
		log.Fatal("Could not assert frame as SynchronisedLyricsFrame")
	}

	// Print the synchronized lyrics.
	fmt.Println("Synchronized Lyrics:")
	for _, text := range sylf.SynchronizedTexts {
		fmt.Printf("[%d] %s\n", text.Timestamp, text.Text)
	}
}
