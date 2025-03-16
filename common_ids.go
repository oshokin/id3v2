package id3v2

import "strings"

// Constants for commonly used ISO 639-2 language codes.
const (
	AlbanianISO6392Code       = "sqi" // ISO 639-2 code for Albanian.
	ArabicISO6392Code         = "ara" // ISO 639-2 code for Arabic.
	BasqueISO6392Code         = "eus" // ISO 639-2 code for Basque.
	BretonISO6392Code         = "bre" // ISO 639-2 code for Breton.
	BulgarianISO6392Code      = "bul" // ISO 639-2 code for Bulgarian.
	CatalanISO6392Code        = "cat" // ISO 639-2 code for Catalan.
	ChineseISO6392Code        = "zho" // ISO 639-2 code for Chinese.
	CornishISO6392Code        = "cor" // ISO 639-2 code for Cornish.
	CroatianISO6392Code       = "hrv" // ISO 639-2 code for Croatian.
	CzechISO6392Code          = "ces" // ISO 639-2 code for Czech.
	DanishISO6392Code         = "dan" // ISO 639-2 code for Danish.
	DutchISO6392Code          = "nld" // ISO 639-2 code for Dutch.
	EnglishISO6392Code        = "eng" // ISO 639-2 code for English.
	EstonianISO6392Code       = "est" // ISO 639-2 code for Estonian.
	FinnishISO6392Code        = "fin" // ISO 639-2 code for Finnish.
	FrenchISO6392Code         = "fra" // ISO 639-2 code for French.
	GermanISO6392Code         = "ger" // ISO 639-2 code for German.
	GreekISO6392Code          = "ell" // ISO 639-2 code for Greek.
	HebrewISO6392Code         = "heb" // ISO 639-2 code for Hebrew.
	HindiISO6392Code          = "hin" // ISO 639-2 code for Hindi.
	HungarianISO6392Code      = "hun" // ISO 639-2 code for Hungarian.
	IcelandicISO6392Code      = "isl" // ISO 639-2 code for Icelandic.
	IndonesianISO6392Code     = "ind" // ISO 639-2 code for Indonesian.
	IrishISO6392Code          = "gle" // ISO 639-2 code for Irish.
	ItalianISO6392Code        = "ita" // ISO 639-2 code for Italian.
	JapaneseISO6392Code       = "jpn" // ISO 639-2 code for Japanese.
	KoreanISO6392Code         = "kor" // ISO 639-2 code for Korean.
	LatvianISO6392Code        = "lav" // ISO 639-2 code for Latvian.
	LithuanianISO6392Code     = "lit" // ISO 639-2 code for Lithuanian.
	MacedonianISO6392Code     = "mkd" // ISO 639-2 code for Macedonian.
	MalayISO6392Code          = "msa" // ISO 639-2 code for Malay.
	MalteseISO6392Code        = "mlt" // ISO 639-2 code for Maltese.
	ManxISO6392Code           = "glv" // ISO 639-2 code for Manx.
	NorwegianISO6392Code      = "nor" // ISO 639-2 code for Norwegian.
	PolishISO6392Code         = "pol" // ISO 639-2 code for Polish.
	PortugueseISO6392Code     = "por" // ISO 639-2 code for Portuguese.
	RomanianISO6392Code       = "ron" // ISO 639-2 code for Romanian.
	RussianISO6392Code        = "rus" // ISO 639-2 code for Russian.
	ScottishGaelicISO6392Code = "gla" // ISO 639-2 code for Scottish Gaelic.
	SerbianISO6392Code        = "srp" // ISO 639-2 code for Serbian.
	SlovakISO6392Code         = "slk" // ISO 639-2 code for Slovak.
	SlovenianISO6392Code      = "slv" // ISO 639-2 code for Slovenian.
	SpanishISO6392Code        = "spa" // ISO 639-2 code for Spanish.
	SwedishISO6392Code        = "swe" // ISO 639-2 code for Swedish.
	ThaiISO6392Code           = "tha" // ISO 639-2 code for Thai.
	TurkishISO6392Code        = "tur" // ISO 639-2 code for Turkish.
	UkrainianISO6392Code      = "ukr" // ISO 639-2 code for Ukrainian.
	VietnameseISO6392Code     = "vie" // ISO 639-2 code for Vietnamese.
	WelshISO6392Code          = "cym" // ISO 639-2 code for Welsh.
)

// Constants for commonly used frame descriptions and IDs.
const (
	ArtistFrameDescription    = "Artist" // Description for the artist frame.
	SubtitleRefinementFrameID = "TIT3"   // ID for the subtitle refinement frame.
	TitleFrameDescription     = "Title"  // Description for the title frame.
	TitleFrameID              = "TIT2"   // ID for the title frame.
	UserDefinedTextFrameID    = "TXXX"   // ID for user-defined text frames.
)

// Common IDs for ID3v2.3 and ID3v2.4.
var (
	// V23CommonIDs maps human-readable descriptions to their corresponding frame IDs in ID3v2.3.
	// For example, "Album/Movie/Show title" maps to "TALB".
	V23CommonIDs = map[string]string{
		"Album/Movie/Show title":         "TALB",
		"Attached picture":               "APIC",
		"Band/Orchestra/Accompaniment":   "TPE2",
		"BPM":                            "TBPM",
		"Chapters":                       "CHAP",
		"Comments":                       "COMM",
		"Composer":                       "TCOM",
		"Conductor/performer refinement": "TPE3",
		"Content group description":      "TIT1",
		"Content type":                   "TCON",
		"Copyright message":              "TCOP",
		"Date":                           "TDAT",
		"Encoded by":                     "TENC",
		"File owner/licensee":            "TOWN",
		"File type":                      "TFLT",
		"Initial key":                    "TKEY",
		"Internet radio station name":    "TRSN",
		"Internet radio station owner":   "TRSO",
		"Interpreted, remixed, or otherwise modified by": "TPE4",
		"ISRC":     "TSRC",
		"Language": "TLAN",
		"Lead artist/Lead performer/Soloist/Performing group": "TPE1",
		"Length":                          "TLEN",
		"Lyricist/Text writer":            "TEXT",
		"Media type":                      "TMED",
		"Original album/movie/show title": "TOAL",
		"Original artist/performer":       "TOPE",
		"Original filename":               "TOFN",
		"Original lyricist/text writer":   "TOLY",
		"Original release year":           "TORY",
		"Part of a set":                   "TPOS",
		"Playlist delay":                  "TDLY",
		"Popularimeter":                   "POPM",
		"Publisher":                       "TPUB",
		"Recording dates":                 "TRDA",
		"Size":                            "TSIZ",
		"Software/Hardware and settings used for encoding": "TSSE",
		"Subtitle/Description refinement":                  SubtitleRefinementFrameID,
		"Synchronised lyrics/text":                         "SYLT",
		"Time":                                             "TIME",
		"Title/Songname/Content description":               TitleFrameID,
		"Track number/Position in set":                     "TRCK",
		"Unique file identifier":                           "UFID",
		"Unsynchronised lyrics/text transcription":         "USLT",
		"User defined text information frame":              UserDefinedTextFrameID,
		"Year":                                             "TYER",

		// Convenience mappings for commonly used frames.
		ArtistFrameDescription: "TPE1",       // Maps "Artist" to "TPE1".
		"Genre":                "TCON",       // Maps "Genre" to "TCON".
		"Title":                TitleFrameID, // Maps "Title" to "TIT2".
	}

	// V24CommonIDs maps human-readable descriptions to their corresponding frame IDs in ID3v2.4.
	// This includes additional frames and updated mappings for ID3v2.4.
	V24CommonIDs = map[string]string{
		"Album sort order":               "TSOA",
		"Album/Movie/Show title":         "TALB",
		"Attached picture":               "APIC",
		"Band/Orchestra/Accompaniment":   "TPE2",
		"BPM":                            "TBPM",
		"Chapters":                       "CHAP",
		"Comments":                       "COMM",
		"Composer":                       "TCOM",
		"Conductor/performer refinement": "TPE3",
		"Content group description":      "TIT1",
		"Content type":                   "TCON",
		"Copyright message":              "TCOP",
		"Encoded by":                     "TENC",
		"Encoding time":                  "TDEN",
		"File owner/licensee":            "TOWN",
		"File type":                      "TFLT",
		"Initial key":                    "TKEY",
		"Internet radio station name":    "TRSN",
		"Internet radio station owner":   "TRSO",
		"Interpreted, remixed, or otherwise modified by": "TPE4",
		"Involved people list":                           "TIPL",
		"ISRC":                                           "TSRC",
		"Language":                                       "TLAN",
		"Lead artist/Lead performer/Soloist/Performing group": "TPE1",
		"Length":                          "TLEN",
		"Lyricist/Text writer":            "TEXT",
		"Media type":                      "TMED",
		"Mood":                            "TMOO",
		"Musician credits list":           "TMCL",
		"Original album/movie/show title": "TOAL",
		"Original artist/performer":       "TOPE",
		"Original filename":               "TOFN",
		"Original lyricist/text writer":   "TOLY",
		"Original release time":           "TDOR",
		"Part of a set":                   "TPOS",
		"Performer sort order":            "TSOP",
		"Playlist delay":                  "TDLY",
		"Popularimeter":                   "POPM",
		"Produced notice":                 "TPRO",
		"Publisher":                       "TPUB",
		"Recording time":                  "TDRC",
		"Release time":                    "TDRL",
		"Set subtitle":                    "TSST",
		"Software/Hardware and settings used for encoding": "TSSE",
		"Subtitle/Description refinement":                  SubtitleRefinementFrameID,
		"Synchronised lyrics/text":                         "SYLT",
		"Tagging time":                                     "TDTG",
		"Title sort order":                                 "TSOT",
		"Title/Songname/Content description":               TitleFrameID,
		"Track number/Position in set":                     "TRCK",
		"Unique file identifier":                           "UFID",
		"Unsynchronised lyrics/text transcription":         "USLT",
		"User defined text information frame":              UserDefinedTextFrameID,

		// Deprecated frames from ID3v2.3, mapped to their ID3v2.4 equivalents.
		"Date":                  "TDRC",
		"Original release year": "TDOR",
		"Recording dates":       "TDRC",
		"Size":                  "",
		"Time":                  "TDRC",
		"Year":                  "TDRC",

		// Convenience mappings for commonly used frames.
		ArtistFrameDescription: "TPE1",       // Maps "Artist" to "TPE1".
		"Genre":                "TCON",       // Maps "Genre" to "TCON".
		"Title":                TitleFrameID, // Maps "Title" to "TIT2".
	}
)

// parsers is a map where the key is the frame ID and the value is a function
// for parsing the corresponding frame. Note that there is no dedicated parser
// for text frames (frames starting with "T"), so you should check for text frames
// explicitly before using this map:
//
//	if strings.HasPrefix(id, "T") {
//	   ...
//	}
var parsers = map[string]func(*bufferedReader, byte) (Framer, error){
	"APIC":                 parsePictureFrame,              // Parser for picture frames.
	"CHAP":                 parseChapterFrame,              // Parser for chapter frames.
	"COMM":                 parseCommentFrame,              // Parser for comment frames.
	"POPM":                 parsePopularimeterFrame,        // Parser for popularimeter frames.
	"SYLT":                 parseSynchronisedLyricsFrame,   // Parser for synchronized lyrics frames.
	UserDefinedTextFrameID: parseUserDefinedTextFrame,      // Parser for user-defined text frames.
	"UFID":                 parseUFIDFrame,                 // Parser for unique file identifier frames.
	"USLT":                 parseUnsynchronisedLyricsFrame, // Parser for unsynchronized lyrics frames.
}

// mustFrameBeInSequence checks if a frame with the given ID must be added to a sequence.
// Some frames, like text frames (starting with "T"), are not added to sequences
// because they are typically unique. Other frames, like "MCDI" or "ETCO", are also
// excluded from sequences.
func mustFrameBeInSequence(id string) bool {
	// Text frames (starting with "T") are not added to sequences.
	if id != UserDefinedTextFrameID && strings.HasPrefix(id, "T") {
		return false
	}

	// Specific frames that should not be added to sequences.
	switch id {
	case "MCDI", "ETCO", "SYTC", "RVRB", "MLLT", "PCNT", "RBUF", "POSS", "OWNE", "SEEK", "ASPI":
	case "IPLS", "RVAD": // Specific ID3v2.3 frames.
		return false
	}

	// All other frames can be added to sequences.
	return true
}
