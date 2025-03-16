package id3v2

import (
	"bytes"
	"testing"
)

func TestUnknownFramesUniqueIdentifiers(t *testing.T) {
	uf1, _ := parseUnknownFrame(newBufferedReader(new(bytes.Buffer)))
	uf2, _ := parseUnknownFrame(newBufferedReader(new(bytes.Buffer)))

	if uf1.UniqueIdentifier() == uf2.UniqueIdentifier() {
		t.Errorf("Two unknown frames have same unique identifiers, " +
			"but every unknown frame should have completely unique identifier.")
	}
}
