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
	// Unknown frames have a random UniqueIdentifier(), so they must not replace each other.
	switch f.(type) {
	case UnknownFrame, *UnknownFrame:
		s.frames = append(s.frames, f)

		return
	}

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

// seqPool reuses sequence objects. See BenchmarkSequencePool in pools_test.go.
var seqPool = sync.Pool{New: func() any {
	return &sequence{frames: []Framer{}} // Create a new sequence with an empty slice of frames.
}}

// getSequence retrieves a sequence object from the pool or creates a new one if the pool is empty.
func getSequence() *sequence {
	s, _ := seqPool.Get().(*sequence)

	return s
}

// putSequence returns a sequence object to the pool for reuse.
// The whole backing array is cleared so pooled sequences do not keep Framer references alive.
func putSequence(s *sequence) {
	clear(s.frames[:cap(s.frames)])
	s.frames = s.frames[:0]

	seqPool.Put(s)
}
