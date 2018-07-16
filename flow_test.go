package chunkasaur

import (
	"testing"
	"math/rand"
	"time"
	"bytes"
	"io"
)

func TestSingleChunkTransmission(t *testing.T) {

	Size := 100000
	virtualFile1 := newVirtualFile(int64(Size))
	PacketSize := 100
	SymbolCount := Size / PacketSize
	EncodingBuffer := 10

	tx := NewTransmitter()
	sourceFileTxInfo := tx.AddFile("s1", virtualFile1, int64(Size))
	tx.ActivateChunk(Chunk{FileInfo: sourceFileTxInfo, Size: sourceFileTxInfo.Size, Offset: 0, PacketSize: int64(PacketSize)})

	rx := NewReceiver()
	rx.PrepareForReception(sourceFileTxInfo, virtualFile1)

	for i := 0; i <= SymbolCount+EncodingBuffer; i++ {
		rx.Receive(tx.GeneratePacket())
	}
	virtualFile1.Validate(t)
}

type virtualFile struct {
	source      []byte
	destination []byte
	io.ReaderAt
	io.WriterAt
}

func newVirtualFile(size int64) (vf *virtualFile) {
	vf = &virtualFile{
		source:      make([]byte, size),
		destination: make([]byte, size),
	}
	rand.Seed(time.Now().UnixNano())
	rand.Read(vf.source)
	return
}

func (vf *virtualFile) ReadAt(p []byte, off int64) (n int, err error) {
	return bytes.NewReader(vf.source).ReadAt(p, off)
}
func (vf *virtualFile) WriteAt(p []byte, off int64) (n int, err error) {
	copy(vf.destination[int(off):], p)
	return len(p), nil
}
func (vf *virtualFile) Validate(t *testing.T) {
	if !bytes.Equal(vf.source, vf.destination) {
		diffCount := 0
		for _, i := range vf.source {
			if vf.source[i] != vf.destination[i] {
				diffCount++
			}
		}
		t.Error("File data was not equal, diffcount", diffCount)
	}
}
