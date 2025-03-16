package id3v2

// Options defines the settings that influence how the tag is processed.
type Options struct {
	// Parse determines whether the tag should be parsed.
	// If set to true, the tag will be parsed; otherwise, it will be skipped.
	Parse bool

	// ParseFrames specifies which frames should be parsed.
	// For example, setting `ParseFrames: []string{"Artist", "Title"}` will only parse
	// the artist and title frames.
	// You can use either frame IDs (e.g., "TPE1", "TIT2")
	// or descriptions (e.g., "Artist", "Title").
	// If ParseFrames is empty or nil, all frames in the tag will be parsed.
	// This option only takes effect if Parse is true.
	//
	// This is particularly useful for improving performance.
	// For instance, if you only need certain text frames, the library will skip parsing
	// large or irrelevant frames like pictures or unknown frames.
	ParseFrames []string
}
