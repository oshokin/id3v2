package id3v2

import (
	"errors"
	"io"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/bytefmt"
)

// defaultSaveBufferSize defines the size of the buffer used during file operations, such as saving or copying.
// It is set to 128 KB, which is a reasonable size for balancing memory usage and I/O performance.
const defaultSaveBufferSize = 128 * bytefmt.KILOBYTE

// ErrNoFile is returned when a tag operation is attempted on a tag that wasn't initialized with a file.
// For example, if you try to save or close a tag that was created without a file.
var ErrNoFile = errors.New("tag was not initialized with file")

// Tag represents an ID3v2 tag in an MP3 file. It stores all the metadata frames, sequences, and other
// relevant information about the tag. You can use it to read, modify, or create ID3v2 tags.
type Tag struct {
	frames    map[string]Framer    // Stores individual frames by their ID.
	sequences map[string]*sequence // Stores sequences of frames (e.g., multiple pictures or comments).

	defaultEncoding Encoding  // The default text encoding used for text frames.
	reader          io.Reader // The reader for the MP3 file.
	originalSize    int64     // The original size of the tag in bytes.
	version         byte      // The ID3v2 version (e.g., 3 or 4).
}

// AddFrame adds a frame to the tag with the specified ID. If the ID is empty or the frame is nil,
// the function does nothing. For frames that can appear multiple times (e.g., pictures or comments),
// use the specialized methods like AddAttachedPicture, AddCommentFrame or AddUnsynchronisedLyricsFrame.
func (tag *Tag) AddFrame(id string, f Framer) {
	if id == "" || f == nil {
		return
	}

	if mustFrameBeInSequence(id) {
		sequence := tag.sequences[id]
		if sequence == nil {
			sequence = getSequence()
		}

		sequence.AddFrame(f)
		tag.sequences[id] = sequence
	} else {
		tag.frames[id] = f
	}
}

// AddAttachedPicture adds a picture frame (e.g., album art) to the tag.
func (tag *Tag) AddAttachedPicture(pf PictureFrame) {
	tag.AddFrame(tag.CommonID("Attached picture"), pf)
}

// AddChapterFrame adds a chapter frame to the tag. Chapters are used to divide an audio file into sections.
func (tag *Tag) AddChapterFrame(cf ChapterFrame) {
	tag.AddFrame(tag.CommonID("Chapters"), cf)
}

// AddCommentFrame adds a comment frame to the tag. Comments can include a description and text.
func (tag *Tag) AddCommentFrame(cf CommentFrame) {
	tag.AddFrame(tag.CommonID("Comments"), cf)
}

// AddTextFrame creates a text frame with the specified encoding and text, then adds it to the tag.
func (tag *Tag) AddTextFrame(id string, encoding Encoding, text string) {
	tag.AddFrame(id, TextFrame{Encoding: encoding, Text: text})
}

// AddUnsynchronisedLyricsFrame adds an unsynchronized lyrics frame to the tag.
// These frames store lyrics without timing information.
func (tag *Tag) AddUnsynchronisedLyricsFrame(uslf UnsynchronisedLyricsFrame) {
	tag.AddFrame(tag.CommonID("Unsynchronised lyrics/text transcription"), uslf)
}

// AddSynchronisedLyricsFrame adds a synchronized lyrics frame to the tag.
// These frames store lyrics with timing information for synchronization with the audio.
func (tag *Tag) AddSynchronisedLyricsFrame(sylf SynchronisedLyricsFrame) {
	tag.AddFrame(tag.CommonID("Synchronised lyrics/text"), sylf)
}

// AddUserDefinedTextFrame adds a user-defined text frame (TXXX) to the tag.
// These frames allow custom metadata to be stored.
func (tag *Tag) AddUserDefinedTextFrame(udtf UserDefinedTextFrame) {
	tag.AddFrame(tag.CommonID("User defined text information frame"), udtf)
}

// AddUFIDFrame adds a unique file identifier frame (UFID) to the tag.
// These frames store a unique identifier for the file.
func (tag *Tag) AddUFIDFrame(ufid UFIDFrame) {
	tag.AddFrame(tag.CommonID("Unique file identifier"), ufid)
}

// CommonID returns the frame ID corresponding to the given description.
// For example, passing "Title" returns "TIT2".
// If the description isn't found, it returns the description itself.
// All descriptions can be found in the common_ids.go.
func (tag *Tag) CommonID(description string) string {
	var ids map[string]string
	if tag.version == 3 {
		ids = V23CommonIDs
	} else {
		ids = V24CommonIDs
	}

	if id, ok := ids[description]; ok {
		return id
	}

	return description
}

// AllFrames returns a map of all frames in the tag.
// The key is the frame ID, and the value is a slice of frames.
// This is useful for inspecting all metadata in the tag.
func (tag *Tag) AllFrames() map[string][]Framer {
	frames := make(map[string][]Framer)

	for id, f := range tag.frames {
		frames[id] = []Framer{f}
	}

	for id, sequence := range tag.sequences {
		frames[id] = sequence.Frames()
	}

	return frames
}

// DeleteAllFrames removes all frames from the tag.
// This is useful for starting fresh when creating a new tag.
func (tag *Tag) DeleteAllFrames() {
	if tag.frames == nil || len(tag.frames) > 0 {
		tag.frames = make(map[string]Framer)
	}

	if tag.sequences == nil || len(tag.sequences) > 0 {
		for _, s := range tag.sequences {
			putSequence(s)
		}

		tag.sequences = make(map[string]*sequence)
	}
}

// DeleteFrames removes all frames with the specified ID from the tag.
func (tag *Tag) DeleteFrames(id string) {
	delete(tag.frames, id)

	if s, ok := tag.sequences[id]; ok {
		putSequence(s)
		delete(tag.sequences, id)
	}
}

// Reset clears all frames in the tag and re-parses the provided reader with the given options.
// This is useful for reusing a tag instance.
func (tag *Tag) Reset(rd io.Reader, opts Options) error {
	tag.DeleteAllFrames()

	return tag.parse(rd, opts)
}

// GetFrames returns all frames with the specified ID.
// If no frames exist, it returns nil.
func (tag *Tag) GetFrames(id string) []Framer {
	if f, exists := tag.frames[id]; exists {
		return []Framer{f}
	} else if s, exists := tag.sequences[id]; exists { //nolint:govet // Shadowing is intentional here.
		return s.Frames()
	}

	return nil
}

// GetLastFrame returns the last frame from the slice returned by GetFrames.
// This is useful for frames that should only appear once, like text frames.
func (tag *Tag) GetLastFrame(id string) Framer {
	// Avoid allocating a slice in GetFrames if there's only one frame.
	if f, exists := tag.frames[id]; exists {
		return f
	}

	fs := tag.GetFrames(id)
	if len(fs) == 0 {
		return nil
	}

	return fs[len(fs)-1]
}

// GetTextFrame returns the text frame with the specified ID.
// If no such frame exists, it returns an empty TextFrame.
func (tag *Tag) GetTextFrame(id string) TextFrame {
	f := tag.GetLastFrame(id)
	if f == nil {
		return TextFrame{}
	}

	tf, _ := f.(TextFrame)

	return tf
}

// DefaultEncoding returns the default text encoding used for text frames in the tag.
func (tag *Tag) DefaultEncoding() Encoding {
	return tag.defaultEncoding
}

// SetDefaultEncoding sets the default text encoding for the tag.
// This encoding is used when adding text frames without explicitly specifying an encoding.
func (tag *Tag) SetDefaultEncoding(encoding Encoding) {
	tag.defaultEncoding = encoding
}

// setDefaultEncodingBasedOnVersion sets the default encoding based on the ID3v2 version.
// ID3v2.4 uses UTF-8 by default, while earlier versions use ISO-8859-1.
func (tag *Tag) setDefaultEncodingBasedOnVersion(version byte) {
	if version == 4 {
		tag.SetDefaultEncoding(EncodingUTF8)
	} else {
		tag.SetDefaultEncoding(EncodingISO)
	}
}

// Count returns the total number of frames in the tag.
func (tag *Tag) Count() int {
	n := len(tag.frames)
	for _, s := range tag.sequences {
		n += s.Count()
	}

	return n
}

// HasFrames checks if the tag contains any frames.
// This is faster than checking Count() > 0.
func (tag *Tag) HasFrames() bool {
	return len(tag.frames) > 0 || len(tag.sequences) > 0
}

// Title returns the title stored in the tag.
func (tag *Tag) Title() string {
	return tag.GetTextFrame(tag.CommonID("Title")).Text
}

// SetTitle sets the title in the tag.
func (tag *Tag) SetTitle(title string) {
	tag.AddTextFrame(tag.CommonID("Title"), tag.DefaultEncoding(), title)
}

// Artist returns the artist stored in the tag.
func (tag *Tag) Artist() string {
	return tag.GetTextFrame(tag.CommonID(ArtistFrameDescription)).Text
}

// SetArtist sets the artist in the tag.
func (tag *Tag) SetArtist(artist string) {
	tag.AddTextFrame(tag.CommonID(ArtistFrameDescription), tag.DefaultEncoding(), artist)
}

// Album returns the album stored in the tag.
func (tag *Tag) Album() string {
	return tag.GetTextFrame(tag.CommonID("Album/Movie/Show title")).Text
}

// SetAlbum sets the album in the tag.
func (tag *Tag) SetAlbum(album string) {
	tag.AddTextFrame(tag.CommonID("Album/Movie/Show title"), tag.DefaultEncoding(), album)
}

// Year returns the year stored in the tag.
func (tag *Tag) Year() string {
	return tag.GetTextFrame(tag.CommonID("Year")).Text
}

// SetYear sets the year in the tag.
func (tag *Tag) SetYear(year string) {
	tag.AddTextFrame(tag.CommonID("Year"), tag.DefaultEncoding(), year)
}

// Genre returns the genre stored in the tag.
func (tag *Tag) Genre() string {
	return tag.GetTextFrame(tag.CommonID("Content type")).Text
}

// SetGenre sets the genre in the tag.
func (tag *Tag) SetGenre(genre string) {
	tag.AddTextFrame(tag.CommonID("Content type"), tag.DefaultEncoding(), genre)
}

// iterateOverAllFrames iterates over every frame in the tag and calls the provided function f.
// This is memory-efficient compared to using AllFrames().
func (tag *Tag) iterateOverAllFrames(f func(id string, frame Framer) error) error {
	for id, frame := range tag.frames {
		if err := f(id, frame); err != nil {
			return err
		}
	}

	for id, sequence := range tag.sequences {
		for _, frame := range sequence.Frames() {
			if err := f(id, frame); err != nil {
				return err
			}
		}
	}

	return nil
}

// Size returns the total size of the tag in bytes, including the tag header and all frames.
func (tag *Tag) Size() int {
	if !tag.HasFrames() {
		return 0
	}

	var n int
	n += tagHeaderSize // Add the size of the tag header.

	err := tag.iterateOverAllFrames(func(_ string, f Framer) error {
		n += frameHeaderSize + f.Size() // Add the size of each frame.

		return nil
	})
	if err != nil {
		panic(err)
	}

	return n
}

// Version returns the ID3v2 version of the tag (e.g., 3 or 4).
func (tag *Tag) Version() byte {
	return tag.version
}

// SetVersion sets the ID3v2 version of the tag.
// If the version is invalid (less than 3 or greater than 4), the function does nothing.
func (tag *Tag) SetVersion(version byte) {
	if version < 3 || version > 4 {
		return
	}

	tag.version = version
	tag.setDefaultEncodingBasedOnVersion(version)
}

// Save writes the tag to the file if the tag was initialized with a file.
// If there are no frames, it writes only the music part without any ID3v2 information.
// Returns ErrNoFile if the tag wasn't initialized with a file.
func (tag *Tag) Save() error {
	file, ok := tag.reader.(*os.File)
	if !ok {
		return ErrNoFile
	}

	// Get the original file's mode (permissions).
	originalFile := file

	originalStat, err := originalFile.Stat()
	if err != nil {
		return err
	}

	// Create a temporary file to write the new tag.
	name := file.Name() + "-id3v2"

	newFile, err := os.OpenFile(filepath.Clean(name), os.O_RDWR|os.O_CREATE, originalStat.Mode())
	if err != nil {
		return err
	}

	// Ensure the temporary file is cleaned up if something goes wrong.
	tempfileShouldBeRemoved := true
	defer func() {
		if tempfileShouldBeRemoved {
			os.Remove(newFile.Name())
		}
	}()

	// Write the tag to the temporary file.
	tagSize, err := tag.WriteTo(newFile)
	if err != nil {
		return err
	}

	// Seek to the music part of the original file.
	if _, err = originalFile.Seek(tag.originalSize, io.SeekStart); err != nil {
		return err
	}

	// Copy the music part to the temporary file.
	buf := getByteSlice(defaultSaveBufferSize)
	defer putByteSlice(buf)

	if _, err = io.CopyBuffer(newFile, originalFile, buf); err != nil {
		return err
	}

	// Close the files to allow replacing.
	newFile.Close()
	originalFile.Close()

	// Replace the original file with the temporary file.
	if err = os.Rename(newFile.Name(), originalFile.Name()); err != nil {
		return err
	}

	tempfileShouldBeRemoved = false

	// Update the tag's reader to the new file.
	tag.reader, err = os.Open(originalFile.Name())
	if err != nil {
		return err
	}

	// Update the tag's original size.
	tag.originalSize = tagSize

	return nil
}

// WriteTo writes the entire tag to the provided writer.
// It returns the number of bytes written and any error encountered.
// If there are no frames, it writes nothing.
func (tag *Tag) WriteTo(w io.Writer) (n int64, err error) {
	if w == nil {
		return 0, errors.New("w is nil")
	}

	// Calculate the size of the frames.
	framesSize := tag.Size() - tagHeaderSize
	if framesSize <= 0 {
		return 0, nil
	}

	// Write the tag header.
	bw := getBufWriter(w)
	defer putBufWriter(bw)

	err = writeTagHeader(bw, uint(framesSize), tag.version)
	if err != nil {
		_ = bw.Flush()

		return int64(bw.Written()), err
	}

	// Write all frames.
	synchSafe := tag.Version() == 4

	err = tag.iterateOverAllFrames(func(id string, f Framer) error {
		return writeFrame(bw, id, f, synchSafe)
	})
	if err != nil {
		_ = bw.Flush()

		return int64(bw.Written()), err
	}

	return int64(bw.Written()), bw.Flush()
}

// writeTagHeader writes the ID3v2 tag header to the provided bufferedWriter.
func writeTagHeader(bw *bufferedWriter, framesSize uint, version byte) error {
	_, err := bw.Write(id3Identifier)
	if err != nil {
		return err
	}

	bw.WriteByte(version)
	bw.WriteByte(0) // Revision
	bw.WriteByte(0) // Flags
	bw.WriteBytesSize(framesSize, true)

	return nil
}

// writeFrame writes a single frame to the provided bufferedWriter.
func writeFrame(bw *bufferedWriter, id string, frame Framer, synchSafe bool) error {
	err := writeFrameHeader(bw, id, truncateIntToUint(frame.Size()), synchSafe)
	if err != nil {
		return err
	}

	_, err = frame.WriteTo(bw)

	return err
}

// writeFrameHeader writes the frame header to the provided bufferedWriter.
func writeFrameHeader(bw *bufferedWriter, id string, frameSize uint, synchSafe bool) error {
	bw.WriteString(id)
	bw.WriteBytesSize(frameSize, synchSafe)

	_, err := bw.Write([]byte{0, 0}) // Flags

	return err
}

// Close closes the tag's file if it was initialized with a file.
// Returns ErrNoFile if the tag wasn't initialized with a file.
func (tag *Tag) Close() error {
	file, ok := tag.reader.(*os.File)
	if !ok {
		return ErrNoFile
	}

	return file.Close()
}
