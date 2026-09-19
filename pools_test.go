package id3v2

import (
	"bytes"
	"io"
	"testing"
)

// Sinks prevent the compiler from eliding work in pool benchmarks.
var (
	sink   []byte
	sinkN  int
	sinkBB *bytes.Buffer
	sinkBR *bufferedReader
	sinkBW *bufferedWriter
	sinkS  *sequence
)

var (
	benchPayload     = make([]byte, 4096)
	benchReadScratch = make([]byte, 64)
	benchSeqFrameA   = CommentFrame{
		Encoding:    EncodingUTF8,
		Language:    EnglishISO6392Code,
		Description: "a",
		Text:        "one",
	}
	benchSeqFrameB = CommentFrame{
		Encoding:    EncodingUTF8,
		Language:    EnglishISO6392Code,
		Description: "b",
		Text:        "two",
	}
)

func TestGetPutByteSlice(t *testing.T) {
	t.Parallel()

	buf := getByteSlice(defaultBufferSize)
	if len(buf) != defaultBufferSize {
		t.Fatalf("len = %d, want %d", len(buf), defaultBufferSize)
	}

	putByteSlice(buf)

	reused := getByteSlice(defaultBufferSize)
	if cap(reused) < defaultBufferSize {
		t.Fatalf("cap = %d, want at least %d", cap(reused), defaultBufferSize)
	}

	putByteSlice(reused)
}

func BenchmarkByteSliceMake32K(b *testing.B) {
	b.ReportAllocs()

	for b.Loop() {
		sink = make([]byte, defaultBufferSize)
	}
}

func BenchmarkByteSlicePool32K(b *testing.B) {
	b.ReportAllocs()

	for b.Loop() {
		buf := getByteSlice(defaultBufferSize)
		putByteSlice(buf)
		sink = buf
	}
}

func BenchmarkByteSliceMake128K(b *testing.B) {
	b.ReportAllocs()

	for b.Loop() {
		sink = make([]byte, defaultSaveBufferSize)
	}
}

func BenchmarkByteSlicePool128K(b *testing.B) {
	b.ReportAllocs()

	for b.Loop() {
		buf := getByteSlice(defaultSaveBufferSize)
		putByteSlice(buf)
		sink = buf
	}
}

func BenchmarkSequenceMake(b *testing.B) {
	b.ReportAllocs()

	for b.Loop() {
		s := &sequence{frames: []Framer{}}
		s.AddFrame(benchSeqFrameA)
		s.AddFrame(benchSeqFrameB)
		sinkS = s
	}
}

func BenchmarkSequencePool(b *testing.B) {
	b.ReportAllocs()

	for b.Loop() {
		s := getSequence()
		s.AddFrame(benchSeqFrameA)
		s.AddFrame(benchSeqFrameB)
		putSequence(s)
		sinkS = s
	}
}

func BenchmarkBytesBufferMake(b *testing.B) {
	b.ReportAllocs()

	for b.Loop() {
		buf := new(bytes.Buffer)
		_, _ = buf.Write(benchPayload)
		sinkN = buf.Len()
		sinkBB = buf
	}
}

func BenchmarkBytesBufferPool(b *testing.B) {
	b.ReportAllocs()

	for b.Loop() {
		buf := getBytesBuffer()
		_, _ = buf.Write(benchPayload)
		sinkN = buf.Len()
		putBytesBuffer(buf)
		sinkBB = buf
	}
}

func BenchmarkBufferedReaderMake(b *testing.B) {
	b.ReportAllocs()

	rd := bytes.NewReader(benchPayload)

	for b.Loop() {
		rd.Reset(benchPayload)

		br := newBufferedReader(rd)
		_, _ = br.Read(benchReadScratch)
		sinkBR = br
	}
}

func BenchmarkBufferedReaderPool(b *testing.B) {
	b.ReportAllocs()

	rd := bytes.NewReader(benchPayload)

	for b.Loop() {
		rd.Reset(benchPayload)

		br := getBufReader(rd)
		_, _ = br.Read(benchReadScratch)
		putBufReader(br)
		sinkBR = br
	}
}

func BenchmarkBufferedWriterMake(b *testing.B) {
	b.ReportAllocs()

	for b.Loop() {
		bw := newBufferedWriter(io.Discard)
		_, _ = bw.Write(benchPayload)
		_ = bw.Flush()
		sinkBW = bw
	}
}

func BenchmarkBufferedWriterPool(b *testing.B) {
	b.ReportAllocs()

	for b.Loop() {
		bw := getBufWriter(io.Discard)
		_, _ = bw.Write(benchPayload)
		_ = bw.Flush()
		putBufWriter(bw)
		sinkBW = bw
	}
}
