package pump

import (
	"bytes"
	"github.com/davecgh/go-spew/spew"
	"io"
	"math/rand"
	"testing"
	"time"
)

func TestSingleChunkMultiFileTransmission(t *testing.T) {
	Size := 1000000    // 1MB
	PacketSize := 1000 // 1KB
	virtualFile1 := newVirtualFile("f1", int64(Size))
	virtualFile2 := newVirtualFile("f2", int64(Size))

	tx := NewTransmitter()
	sourceFileTxInfo1 := tx.AddObject("s1", virtualFile1, int64(Size))
	sourceFileTxInfo2 := tx.AddObject("s2", virtualFile2, int64(Size))
	tx.ActivateChunk(Chunk{Object: sourceFileTxInfo1, Size: sourceFileTxInfo1.Size, Offset: 0, PacketSize: int64(PacketSize)})
	tx.ActivateChunk(Chunk{Object: sourceFileTxInfo2, Size: sourceFileTxInfo2.Size, Offset: 0, PacketSize: int64(PacketSize)})

	rx := NewReceiver()
	rx.PrepareForReception(sourceFileTxInfo1, virtualFile1)
	rx.PrepareForReception(sourceFileTxInfo2, virtualFile2)

	for !rx.Idle() {
		rx.Receive(tx.GeneratePacket())
	}

	virtualFile1.Validate(t)
	virtualFile2.Validate(t)
}

func TestPaddingOnOddSizedFiles(t *testing.T) {
	Size := 12345
	virtualFile1 := newVirtualFile("f1", int64(Size))
	PacketSize := 89 // 1KB

	tx := NewTransmitter()
	sourceFileTxInfo1 := tx.AddObject("s1", virtualFile1, int64(Size))
	tx.ActivateChunk(Chunk{Object: sourceFileTxInfo1, Size: sourceFileTxInfo1.Size, Offset: 0, PacketSize: int64(PacketSize)})

	rx := NewReceiver()
	rx.PrepareForReception(sourceFileTxInfo1, virtualFile1)

	for !rx.Idle() {
		rx.Receive(tx.GeneratePacket())
	}
	virtualFile1.Validate(t)
}

func TestMultiChunkTransmission(t *testing.T) {

	evenFile := newVirtualFile("even", int64(10000))
	oddFile := newVirtualFile("odd", int64(12345))

	tx := NewTransmitter()
	evenTxInfo := tx.AddObject(evenFile.id, evenFile, evenFile.size())
	oddTxInfo := tx.AddObject(oddFile.id, oddFile, oddFile.size())
	tx.ActivateChunk(Chunk{Object: evenTxInfo, Size: evenFile.size() / 2, Offset: 0, PacketSize: 100})
	tx.ActivateChunk(Chunk{Object: evenTxInfo, Size: evenFile.size() / 2, Offset: evenTxInfo.Size / 2, PacketSize: 100})

	tx.ActivateChunk(Chunk{Object: oddTxInfo, Size: 8392, Offset: 0, PacketSize: 100})
	tx.ActivateChunk(Chunk{Object: oddTxInfo, Size: 3953, Offset: 8392, PacketSize: 100})

	rx := NewReceiver()
	rx.PrepareForReception(evenTxInfo, evenFile)
	rx.PrepareForReception(oddTxInfo, oddFile)

	for !rx.Idle() {
		rx.Receive(tx.GeneratePacket())
	}
	evenFile.Validate(t)
	oddFile.Validate(t)
}

func TestEncodingDemo(t *testing.T) {
	Size := 64
	PacketSize := 16

	virtualFile1 := newVirtualFile("f1", int64(Size))
	spew.Dump(virtualFile1.source)

	tx := NewTransmitter()
	sourceFileTxInfo1 := tx.AddObject("s1", virtualFile1, int64(Size))

	tx.ActivateChunk(Chunk{Object: sourceFileTxInfo1, Size: sourceFileTxInfo1.Size, Offset: 0, PacketSize: int64(PacketSize)})

	rx := NewReceiver()
	rx.PrepareForReception(sourceFileTxInfo1, virtualFile1)

	for !rx.Idle() {
		packet := tx.GeneratePacket()
		spew.Dump(packet.Block.Data)
		rx.Receive(packet)
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

func newVirtualFile(id string, size int64) (vf *virtualTestFile) {
	vf = &virtualTestFile{
		source:      make([]byte, size),
		destination: make([]byte, size),
		id:          id,
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
func (vf *virtualTestFile) size() int64 {
	return int64(len(vf.source))
}
