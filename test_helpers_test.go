package id3v2

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestFixturesPresent(t *testing.T) {
	t.Parallel()

	for _, path := range []string{mp3Path, multiMp3Path, frontCoverPath, backCoverPath} {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("required fixture %s: %v", path, err)
		}
	}
}

func copyFixture(t testing.TB) string {
	t.Helper()

	src, err := os.Open(mp3Path)
	if err != nil {
		t.Fatalf("open fixture %q: %v", mp3Path, err)
	}
	defer src.Close()

	dstPath := filepath.Join(t.TempDir(), filepath.Base(mp3Path))

	dst, err := os.Create(dstPath)
	if err != nil {
		t.Fatalf("create fixture copy: %v", err)
	}

	if _, err := io.Copy(dst, src); err != nil {
		_ = dst.Close()

		t.Fatalf("copy fixture: %v", err)
	}

	if err := dst.Close(); err != nil {
		t.Fatalf("close fixture copy: %v", err)
	}

	return dstPath
}

func assertRoundTrip(t *testing.T, version byte, frameID string, frame Framer) Framer {
	t.Helper()

	tag := NewEmptyTag()
	tag.SetVersion(version)
	tag.AddFrame(frameID, frame)

	var buf bytes.Buffer

	if _, err := tag.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}

	parsed, err := ParseReader(bytes.NewReader(buf.Bytes()), Options{Parse: true})
	if err != nil {
		t.Fatalf("ParseReader: %v", err)
	}

	t.Cleanup(func() {
		_ = parsed.Close()
	})

	frames := parsed.GetFrames(frameID)
	if len(frames) == 0 {
		t.Fatalf("no frames with id %q after round-trip", frameID)
	}

	return frames[len(frames)-1]
}

func assertFrameSizeMatchesWrite(t *testing.T, frame Framer) {
	t.Helper()

	var buf bytes.Buffer

	n, err := frame.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo: %v", err)
	}

	if got, want := int(n), buf.Len(); got != want {
		t.Fatalf("WriteTo n = %d, buffer = %d", got, want)
	}

	if got, want := frame.Size(), buf.Len(); got != want {
		t.Fatalf("Size() = %d, actual = %d", got, want)
	}
}

func assertVersionedFrameSizeMatchesWrite(t *testing.T, frame versionedFramer, version byte) {
	t.Helper()

	var buf bytes.Buffer

	n, err := frame.writeToVersion(&buf, version)
	if err != nil {
		t.Fatalf("writeToVersion(%d): %v", version, err)
	}

	if got, want := int(n), buf.Len(); got != want {
		t.Fatalf("writeToVersion(%d) n = %d, buffer = %d", version, got, want)
	}

	if got, want := frame.sizeForVersion(version), buf.Len(); got != want {
		t.Fatalf("sizeForVersion(%d) = %d, actual = %d", version, got, want)
	}
}

func audioPayload(t *testing.T, path string, originalSize int64) []byte {
	t.Helper()

	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %q: %v", path, err)
	}
	defer f.Close()

	if _, err := f.Seek(originalSize, io.SeekStart); err != nil {
		t.Fatalf("seek audio payload: %v", err)
	}

	data, err := io.ReadAll(f)
	if err != nil {
		t.Fatalf("read audio payload: %v", err)
	}

	return data
}
