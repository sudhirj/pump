package chunkasaur

import (
	"testing"
	"math/rand"
	"time"
	"bytes"
	"io"
)

func TestSingleChunkTransmission(t *testing.T) {
	Size := 1000000 // 1MB
	virtualFile1 := newVirtualFile("f1",int64(Size))
	virtualFile2 := newVirtualFile("f2",int64(Size))
	PacketSize := 1000 // 1KB
	SymbolCount := Size / PacketSize // 1000 packets
	EncodingBuffer := 100

	tx := NewTransmitter()
	sourceFileTxInfo1 := tx.AddFile("s1", virtualFile1, int64(Size))
	sourceFileTxInfo2 := tx.AddFile("s2", virtualFile2, int64(Size))
	tx.ActivateChunk(Chunk{FileInfo: sourceFileTxInfo1, Size: sourceFileTxInfo1.Size, Offset: 0, PacketSize: int64(PacketSize)})
	tx.ActivateChunk(Chunk{FileInfo: sourceFileTxInfo2, Size: sourceFileTxInfo2.Size, Offset: 0, PacketSize: int64(PacketSize)})

	rx := NewReceiver()
	rx.PrepareForReception(sourceFileTxInfo1, virtualFile1)
	rx.PrepareForReception(sourceFileTxInfo2, virtualFile2)

	for i := 0; i <= 2*(SymbolCount+EncodingBuffer); i++ {
		rx.Receive(tx.GeneratePacket())
	}
	virtualFile1.Validate(t)
	virtualFile2.Validate(t)
}

func TestPaddingOnOddSizedFiles(t *testing.T) {
	Size := 12345
	virtualFile1 := newVirtualFile("f1",int64(Size))
	PacketSize := 89 // 1KB
	SymbolCount := Size / PacketSize // 1000 packets
	EncodingBuffer := 100

	tx := NewTransmitter()
	sourceFileTxInfo1 := tx.AddFile("s1", virtualFile1, int64(Size))
	tx.ActivateChunk(Chunk{FileInfo: sourceFileTxInfo1, Size: sourceFileTxInfo1.Size, Offset: 0, PacketSize: int64(PacketSize)})

	rx := NewReceiver()
	rx.PrepareForReception(sourceFileTxInfo1, virtualFile1)

	for i := 0; i <= 1*(SymbolCount+EncodingBuffer); i++ {
		rx.Receive(tx.GeneratePacket())
	}
	virtualFile1.Validate(t)
}

type virtualTestFile struct {
	source      []byte
	destination []byte
	io.ReaderAt
	io.WriterAt
	id string
}

func newVirtualFile(id string,size int64) (vf *virtualTestFile) {
	vf = &virtualTestFile{
		source:      make([]byte, size),
		destination: make([]byte, size),
		id: id,
	}
	rand.Seed(time.Now().UnixNano())
	rand.Read(vf.source)
	return
}

func (vf *virtualTestFile) ReadAt(p []byte, off int64) (n int, err error) {
	return bytes.NewReader(vf.source).ReadAt(p, off)
}
func (vf *virtualTestFile) WriteAt(p []byte, off int64) (n int, err error) {
	return copy(vf.destination[int(off):], p), nil
}
func (vf *virtualTestFile) Validate(t *testing.T) {
	if !bytes.Equal(vf.source, vf.destination) {
		diffCount := 0
		for _, i := range vf.source {
			if vf.source[i] != vf.destination[i] {
				diffCount++
			}
		}
		t.Error(vf.id, "File data was not equal, diffcount", diffCount)
	}
}
