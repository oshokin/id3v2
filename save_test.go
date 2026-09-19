package id3v2

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestSavePreservesAudioPayload(t *testing.T) {
	path := copyFixture(t)

	tag, err := Open(path, Options{Parse: true})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	before := audioPayload(t, path, tag.originalSize)

	tag.SetTitle("Updated title")

	if err := tag.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	if err := tag.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	reopened, err := Open(path, Options{Parse: true})
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer reopened.Close()

	after := audioPayload(t, path, reopened.originalSize)
	if string(before) != string(after) {
		t.Fatalf("audio payload changed after Save")
	}

	if reopened.Title() != "Updated title" {
		t.Fatalf("Title() = %q, want Updated title", reopened.Title())
	}
}

func TestSaveRemovesOldTagBytes(t *testing.T) {
	path := copyFixture(t)

	tag, err := Open(path, Options{Parse: false})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	tag.SetTitle(strings.Repeat("A", 4096))

	if err := tag.Save(); err != nil {
		t.Fatalf("first Save: %v", err)
	}

	if err := tag.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	tag, err = Open(path, Options{Parse: false})
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}

	tag.SetTitle("B")

	if err := tag.Save(); err != nil {
		t.Fatalf("second Save: %v", err)
	}

	if err := tag.Close(); err != nil {
		t.Fatalf("Close after shrink: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}

	reopened, err := Open(path, Options{Parse: true})
	if err != nil {
		t.Fatalf("parse shrunk tag: %v", err)
	}
	defer reopened.Close()

	wantLen := reopened.originalSize + int64(len(audioPayload(t, path, reopened.originalSize)))
	if info.Size() != wantLen {
		t.Fatalf("file size = %d, want %d", info.Size(), wantLen)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	if strings.Contains(string(raw), strings.Repeat("A", 4096)) {
		t.Fatal("old large tag bytes are still present")
	}

	if reopened.Title() != "B" {
		t.Fatalf("Title() = %q, want B", reopened.Title())
	}
}

func TestReplaceFileOverwritesExisting(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst")

	if err := os.WriteFile(src, []byte("new"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(dst, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := replaceFile(src, dst); err != nil {
		t.Fatalf("replaceFile: %v", err)
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}

	if string(got) != "new" {
		t.Fatalf("dst = %q, want new", got)
	}
}

func TestSavePreservesFileMode(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix permission bits are not preserved on Windows")
	}

	path := copyFixture(t)

	if err := os.Chmod(path, 0o640); err != nil {
		t.Fatalf("Chmod: %v", err)
	}

	tag, err := Open(path, Options{Parse: false})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	tag.SetTitle("Mode")

	if err := tag.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	if err := tag.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}

	if got := info.Mode().Perm(); got != 0o640 {
		t.Fatalf("file mode = %o, want 640", got)
	}
}

func TestSaveCanBeCalledMoreThanOnce(t *testing.T) {
	path := copyFixture(t)

	tag, err := Open(path, Options{Parse: false})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer tag.Close()

	tag.SetTitle("First")

	if err := tag.Save(); err != nil {
		t.Fatalf("first Save: %v", err)
	}

	tag.SetArtist("Second")

	if err := tag.Save(); err != nil {
		t.Fatalf("second Save: %v", err)
	}

	reopened, err := Open(path, Options{Parse: true})
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer reopened.Close()

	if reopened.Title() != "First" {
		t.Fatalf("Title() = %q, want First", reopened.Title())
	}

	if reopened.Artist() != "Second" {
		t.Fatalf("Artist() = %q, want Second", reopened.Artist())
	}
}

func TestSaveDoesNotLeaveTemporaryFileOnSuccess(t *testing.T) {
	path := copyFixture(t)

	tag, err := Open(path, Options{Parse: false})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	tag.SetTitle("Temp")

	if err := tag.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	if err := tag.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	assertNoTempFiles(t, filepath.Dir(path))
}

func TestSaveDoesNotLeaveTemporaryFileOnWriteFailure(t *testing.T) {
	path := copyFixture(t)

	tag, err := Open(path, Options{Parse: false})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer tag.Close()

	tag.AddCommentFrame(CommentFrame{
		Encoding: EncodingUTF8,
		Language: "en",
		Text:     "invalid language length",
	})

	if err := tag.Save(); err == nil {
		t.Fatal("Save() succeeded, want write failure")
	}

	assertNoTempFiles(t, filepath.Dir(path))
}

func assertNoTempFiles(t *testing.T, dir string) {
	t.Helper()

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}

	for _, entry := range entries {
		if strings.Contains(entry.Name(), ".id3v2-") {
			t.Fatalf("temporary file left behind: %s", entry.Name())
		}
	}
}
