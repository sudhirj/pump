package chunkasaur

import (
	"io"
	"github.com/google/gofountain"
)

type Transmitter struct {
	readers           map[FileInfo]io.ReaderAt
	chunks            map[ChunkInfo][]fountain.LTBlock
	packetIndex       int64
	chunkBlockIndexes map[ChunkInfo]int64
}

func NewTransmitter() *Transmitter {
	return &Transmitter{
		readers:           make(map[FileInfo]io.ReaderAt),
		chunks:            make(map[ChunkInfo][]fountain.LTBlock),
		chunkBlockIndexes: make(map[ChunkInfo]int64),
	}
}

func (tx *Transmitter) AddFile(id string, r io.ReaderAt, fileSize uint64) (fd FileInfo) {
	fd.ID = id
	fd.Size = fileSize
	tx.readers[fd] = r
	return
}
func (tx *Transmitter) ActivateChunkWithWeight(cd ChunkInfo, weight int) {
	data := make([]byte, cd.Size)                          // Set up a buffer with chunk size
	tx.readers[cd.FileInfo].ReadAt(data, int64(cd.Offset)) // and read that data from the file
	ids := buildIds(cd.TargetBlockCount())
	tx.chunks[cd] = fountain.EncodeLTBlocks(data, ids, cd.Codec())
}

func buildIds(count int64) []int64 {
	ids := make([]int64, count)
	for i := 0; i < len(ids); i++ {
		ids[i] = int64(i)
	}
	return ids
}
func (tx *Transmitter) GeneratePacket() (pck Packet) {
	chosenChunk := tx.chooseChunk()
	chosenBlockIndex := tx.chooseBlockIndex(chosenChunk)
	return Packet{
		ChunkInfo: chosenChunk,
		Block:     tx.chunks[chosenChunk][chosenBlockIndex],
	}
}

func (tx *Transmitter) ActivateChunk(cd ChunkInfo)   { tx.ActivateChunkWithWeight(cd, 1) }
func (tx *Transmitter) DeactivateChunk(cd ChunkInfo) { tx.ActivateChunkWithWeight(cd, 0) }
func (tx *Transmitter) chooseChunk() ChunkInfo {
	allActiveChunks := make([]ChunkInfo, len(tx.chunks))
	i := 0
	for c := range tx.chunks {
		allActiveChunks[i] = c
	}
	idx := tx.packetIndex % int64(len(allActiveChunks))
	tx.packetIndex++
	return allActiveChunks[idx]
}
func (tx *Transmitter) chooseBlockIndex(cd ChunkInfo) int64 {
	idx := tx.chunkBlockIndexes[cd]
	tx.chunkBlockIndexes[cd]++
	return idx % int64(len(tx.chunks[cd]))

}

type Receiver struct {
	writers        map[FileInfo]io.WriterAt
	chunkDecoders  map[ChunkInfo]fountain.Decoder
	finishedChunks map[ChunkInfo]struct{}
}

func NewReceiver() *Receiver {
	return &Receiver{
		writers:        make(map[FileInfo]io.WriterAt),
		chunkDecoders:  make(map[ChunkInfo]fountain.Decoder),
		finishedChunks: make(map[ChunkInfo]struct{}),
	}
}

func (rx *Receiver) PrepareForReception(fd FileInfo, w io.WriterAt) {
	rx.writers[fd] = w
}
func (rx *Receiver) Receive(pck Packet) {
	if _, done := rx.finishedChunks[pck.ChunkInfo]; done {
		return
	}
	if _, present := rx.chunkDecoders[pck.ChunkInfo]; !present {
		rx.chunkDecoders[pck.ChunkInfo] = pck.ChunkInfo.Decoder()
	}

	if rx.chunkDecoders[pck.ChunkInfo].AddBlocks([]fountain.LTBlock{pck.Block}) {
		rx.writers[pck.ChunkInfo.FileInfo].WriteAt(rx.chunkDecoders[pck.ChunkInfo].Decode(), int64(pck.ChunkInfo.Offset))
		rx.finishedChunks[pck.ChunkInfo] = struct{}{}
		delete(rx.chunkDecoders, pck.ChunkInfo)
	}
}

type FileInfo struct {
	ID   string
	Size uint64
}

type ChunkInfo struct {
	FileInfo   FileInfo
	Size       uint64
	Offset     uint64
	PacketSize uint64
}

func (c ChunkInfo) SourceBlockCount() int64 {
	return int64(float64(c.Size / c.PacketSize))
}
func (c ChunkInfo) Decoder() fountain.Decoder {
	return fountain.NewRaptorCodec(int(c.SourceBlockCount()), 8).NewDecoder(int(c.Size))
}
func (c ChunkInfo) Codec() fountain.Codec {
	return fountain.NewRaptorCodec(int(c.SourceBlockCount()), 8)
}
func (c ChunkInfo) TargetBlockCount() int64 {
	return int64(float64(c.SourceBlockCount()) * 1.1)
}

type Packet struct {
	ChunkInfo ChunkInfo
	Block     fountain.LTBlock
}
