package id3v2

import (
	"sync"
)

// sequence is a structure used to manage frames that can appear multiple times in an ID3v2 tag.
// Examples of such frames include APIC (attached pictures), COMM (comments), and USLT (unsynchronized lyrics).
// This structure ensures that frames with the same unique identifier are not duplicated.
type sequence struct {
	frames []Framer // frames holds a slice of Framer interfaces representing the frames in the sequence.
}

// AddFrame adds a frame to the sequence. If a frame with the same unique identifier already exists,
// it replaces the existing frame. Otherwise, it appends the new frame to the sequence.
func (s *sequence) AddFrame(f Framer) {
	i := indexOfFrame(f, s.frames) // Find the index of the frame with the same unique identifier.

	if i == -1 {
		// If the frame doesn't exist in the sequence, append it.
		s.frames = append(s.frames, f)
	} else {
		// If the frame already exists, replace it with the new one.
		s.frames[i] = f
	}
}

// indexOfFrame searches for a frame in the given slice of frames and returns its index.
// It uses the frame's unique identifier to determine if two frames are the same.
// If the frame is not found, it returns -1.
func indexOfFrame(f Framer, fs []Framer) int {
	for i, ff := range fs {
		if f.UniqueIdentifier() == ff.UniqueIdentifier() {
			return i // Return the index if the frame is found.
		}
	}

	return -1 // Return -1 if the frame is not found.
}

// Count returns the number of frames in the sequence.
func (s *sequence) Count() int {
	return len(s.frames)
}

// Frames returns a slice of all frames in the sequence.
func (s *sequence) Frames() []Framer {
	return s.frames
}

// seqPool is a sync.Pool used to reuse sequence objects to reduce memory allocations.
// This improves performance by avoiding frequent creation and garbage collection of sequence objects.
var seqPool = sync.Pool{New: func() any {
	return &sequence{frames: []Framer{}} // Create a new sequence with an empty slice of frames.
}}

// getSequence retrieves a sequence object from the pool or creates a new one if the pool is empty.
// If the retrieved sequence has existing frames, it resets the frames slice to ensure a clean state.
func getSequence() *sequence {
	s, _ := seqPool.Get().(*sequence) // Retrieve a sequence from the pool.
	if s.Count() > 0 {
		s.frames = []Framer{} // Reset the frames slice if it contains any frames.
	}

	return s
}

// putSequence returns a sequence object to the pool for reuse.
// This helps reduce memory allocations by reusing sequence objects instead of discarding them.
func putSequence(s *sequence) {
	seqPool.Put(s) // Return the sequence to the pool.
}
