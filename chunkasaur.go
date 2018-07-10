package chunkasaur

import (
	"io"
	"github.com/google/gofountain"
)

type Transmitter struct {
	readers           map[FileInfo]io.ReaderAt
	chunkBlocks       map[Chunk][]fountain.LTBlock
	packetIndex       int64
	chunkBlockIndexes map[Chunk]int64
}

func NewTransmitter() *Transmitter {
	return &Transmitter{
		readers:           make(map[FileInfo]io.ReaderAt),
		chunkBlocks:       make(map[Chunk][]fountain.LTBlock),
		chunkBlockIndexes: make(map[Chunk]int64),
	}
}

func (tx *Transmitter) AddFile(id string, r io.ReaderAt, fileSize int64) (f FileInfo) {
	f.ID = id
	f.Size = fileSize
	tx.readers[f] = r
	return
}
func (tx *Transmitter) ActivateChunkWithWeight(chunk Chunk, weight int) {
	data := make([]byte, chunk.Size)                      // Set up a buffer with chunk size
	tx.readers[chunk.FileInfo].ReadAt(data, chunk.Offset) // and read that data from the file
	ids := buildIds(chunk.targetBlockCount())
	tx.chunkBlocks[chunk] = fountain.EncodeLTBlocks(data, ids, chunk.codec())
}
func (tx *Transmitter) GeneratePacket() (packet Packet) {
	chosenChunk := tx.chooseChunk()
	chosenBlockIndex := tx.chooseBlockIndex(chosenChunk)
	return Packet{
		Chunk: chosenChunk,
		Block: tx.chunkBlocks[chosenChunk][chosenBlockIndex],
	}
}
func (tx *Transmitter) ActivateChunk(chunk Chunk)   { tx.ActivateChunkWithWeight(chunk, 1) }
func (tx *Transmitter) DeactivateChunk(chunk Chunk) {}

func (tx *Transmitter) chooseChunk() Chunk {
	var allActiveChunks []Chunk
	for c := range tx.chunkBlocks { // Not optimal, but good enough since N is usually small
		allActiveChunks = append(allActiveChunks, c)
	}
	idx := tx.packetIndex % int64(len(allActiveChunks))
	tx.packetIndex++
	return allActiveChunks[idx]
}
func (tx *Transmitter) chooseBlockIndex(chunk Chunk) int64 {
	idx := tx.chunkBlockIndexes[chunk] % int64(len(tx.chunkBlocks[chunk]))
	tx.chunkBlockIndexes[chunk]++
	return idx

}
func buildIds(count int64) []int64 {
	ids := make([]int64, count)
	for i := 0; i < len(ids); i++ {
		ids[i] = int64(i)
	}
	return ids
}

type Receiver struct {
	writers        map[FileInfo]io.WriterAt
	chunkDecoders  map[Chunk]fountain.Decoder
	finishedChunks map[Chunk]struct{}
}

func NewReceiver() *Receiver {
	return &Receiver{
		writers:        make(map[FileInfo]io.WriterAt),
		chunkDecoders:  make(map[Chunk]fountain.Decoder),
		finishedChunks: make(map[Chunk]struct{}),
	}
}

func (rx *Receiver) PrepareForReception(f FileInfo, w io.WriterAt) {
	rx.writers[f] = w
}
func (rx *Receiver) Receive(packet Packet) {
	if _, done := rx.finishedChunks[packet.Chunk]; done {
		return
	}
	if _, present := rx.chunkDecoders[packet.Chunk]; !present {
		rx.chunkDecoders[packet.Chunk] = packet.Chunk.decoder()
	}

	if rx.chunkDecoders[packet.Chunk].AddBlocks([]fountain.LTBlock{packet.Block}) {
		rx.writers[packet.Chunk.FileInfo].WriteAt(rx.chunkDecoders[packet.Chunk].Decode(), packet.Chunk.Offset)
		rx.finishedChunks[packet.Chunk] = struct{}{}
		delete(rx.chunkDecoders, packet.Chunk)
	}
}

type FileInfo struct {
	ID   string
	Size int64
}

type Chunk struct {
	FileInfo   FileInfo
	Size       int64
	Offset     int64
	PacketSize int64
}

func (c Chunk) sourceBlockCount() int64 {
	return int64(float64(c.Size / c.PacketSize))
}
func (c Chunk) decoder() fountain.Decoder {
	return fountain.NewRaptorCodec(int(c.sourceBlockCount()), 8).NewDecoder(int(c.Size))
}
func (c Chunk) codec() fountain.Codec {
	return fountain.NewRaptorCodec(int(c.sourceBlockCount()), 8)
}
func (c Chunk) targetBlockCount() int64 {
	return c.sourceBlockCount() + 5
}

type Packet struct {
	Chunk Chunk
	Block fountain.LTBlock
}
