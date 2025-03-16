package id3v2

import (
	"io"
	"os"
	"path/filepath"
)

// Available picture types for picture frames (APIC frames).
// These constants define the type of image stored in the picture frame.
const (
	PTOther                   = iota // Other type of image
	PTFileIcon                       // 32x32 PNG file icon
	PTOtherFileIcon                  // Other file icon
	PTFrontCover                     // Front cover image (e.g., album art)
	PTBackCover                      // Back cover image
	PTLeafletPage                    // Leaflet page (e.g., booklet)
	PTMedia                          // Media image (e.g., CD label)
	PTLeadArtistSoloist              // Lead artist or soloist image
	PTArtistPerformer                // Artist or performer image
	PTConductor                      // Conductor image
	PTBandOrchestra                  // Band or orchestra image
	PTComposer                       // Composer image
	PTLyricistTextWriter             // Lyricist or text writer image
	PTRecordingLocation              // Recording location image
	PTDuringRecording                // Image captured during recording
	PTDuringPerformance              // Image captured during performance
	PTMovieScreenCapture             // Movie or video screen capture
	PTBrightColouredFish             // Bright-colored fish image (used for testing)
	PTIllustration                   // Illustration image
	PTBandArtistLogotype             // Band or artist logotype
	PTPublisherStudioLogotype        // Publisher or studio logotype
)

// Open opens the MP3 file specified by `name` and parses its ID3v2 tag.
// If the file does not contain an ID3v2 tag, a new one is created with ID3v2.4 version.
// The `opts` parameter controls parsing behavior, such as whether to parse all frames or specific ones.
// Returns a pointer to the Tag and an error if the file cannot be opened or parsed.
func Open(name string, opts Options) (*Tag, error) {
	// Open the file and clean the path to prevent directory traversal issues.
	file, err := os.Open(filepath.Clean(name))
	if err != nil {
		return nil, err
	}

	// Parse the file's content using ParseReader.
	return ParseReader(file, opts)
}

// ParseReader reads from the provided `io.Reader` and parses the ID3v2 tag.
// If no tag is found, a new ID3v2.4 tag is created.
// The `opts` parameter controls parsing behavior, such as whether to parse all frames or specific ones.
// Returns a pointer to the Tag and an error if parsing fails.
func ParseReader(rd io.Reader, opts Options) (*Tag, error) {
	// Create a new empty tag and parse the reader's content into it.
	tag := NewEmptyTag()
	err := tag.parse(rd, opts)

	return tag, err
}

// NewEmptyTag creates and returns a new empty ID3v2.4 tag.
// The tag has no frames and no associated reader.
// This is useful for creating a new tag from scratch.
func NewEmptyTag() *Tag {
	// Create a new Tag instance and initialize it with default values.
	tag := new(Tag)
	tag.init(nil, 0, 4) // Initialize with no reader, zero size, and ID3v2.4 version.

	return tag
}
