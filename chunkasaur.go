package chunkasaur

import "io"

type Transmitter struct {
}

func (tx *Transmitter) AddFile(id string, r io.ReaderAt, fileSize int64) (fd File) { return }
func (tx *Transmitter) ActivateChunk(cd Chunk)                                     { tx.ActivateChunkWithWeight(cd, 1) }
func (tx *Transmitter) ActivateChunkWithWeight(cd Chunk, weight int)               {}
func (tx *Transmitter) DeactivateChunk(cd Chunk)                                   { tx.ActivateChunkWithWeight(cd, 0) }
func (tx *Transmitter) GeneratePacket() (pck Packet)                               { return }

type Receiver struct {
}

func (rx *Receiver) Expect(fd File, localPath string) {}
func (rx *Receiver) Receive(pck Packet)               {}

type File struct {
	ID   string
	Size uint64
}

type Chunk struct {
	File       File
	Size       uint64
	Offset     uint64
	PacketSize uint64
}

type Packet struct {
	Chunk Chunk
	Data  []byte
}
