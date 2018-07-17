package chunkasaur

import (
	"io"
	"github.com/google/gofountain"
	"math"
)

type Transmitter struct {
	readers           map[Object]io.ReaderAt
	chunkBlocks       map[Chunk][]fountain.LTBlock
	packetIndex       int64
	chunkBlockIndexes map[Chunk]int64
}

func NewTransmitter() *Transmitter {
	return &Transmitter{
		readers:           make(map[Object]io.ReaderAt),
		chunkBlocks:       make(map[Chunk][]fountain.LTBlock),
		chunkBlockIndexes: make(map[Chunk]int64),
	}
}

func (tx *Transmitter) AddFile(id string, r io.ReaderAt, fileSize int64) (o Object) {
	o.ID = id
	o.Size = fileSize
	tx.readers[o] = r
	return
}
func (tx *Transmitter) ActivateChunkWithWeight(chunk Chunk, weight int) {
	data := make([]byte, chunk.Size)
	tx.readers[chunk.FileInfo].ReadAt(data, chunk.Offset)
	tx.chunkBlocks[chunk] = chunk.encode(data)
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
	idx := tx.packetIndex % int64(len(tx.chunkBlocks))
	tx.packetIndex++
	return tx.activeChunks()[idx]
}
func (tx *Transmitter) activeChunks() (activeChunks []Chunk) {
	for c := range tx.chunkBlocks { // Not optimal, but good enough since N is usually small
		activeChunks = append(activeChunks, c)
	}
	return
}
func (tx *Transmitter) chooseBlockIndex(chunk Chunk) int64 {
	idx := tx.chunkBlockIndexes[chunk] % int64(len(tx.chunkBlocks[chunk]))
	tx.chunkBlockIndexes[chunk]++
	return idx

}

type Receiver struct {
	writers        map[Object]io.WriterAt
	chunkDecoders  map[Chunk]fountain.Decoder
	finishedChunks map[Chunk]struct{}
}

func NewReceiver() *Receiver {
	return &Receiver{
		writers:        make(map[Object]io.WriterAt),
		chunkDecoders:  make(map[Chunk]fountain.Decoder),
		finishedChunks: make(map[Chunk]struct{}),
	}
}

func (rx *Receiver) PrepareForReception(o Object, w io.WriterAt) {
	rx.writers[o] = w
}
func (rx *Receiver) Receive(packet Packet) {
	if _, done := rx.finishedChunks[packet.Chunk]; done {
		return
	}
	if _, present := rx.chunkDecoders[packet.Chunk]; !present {
		rx.chunkDecoders[packet.Chunk] = packet.Chunk.decoder()
	}

	if rx.chunkDecoders[packet.Chunk].AddBlocks([]fountain.LTBlock{packet.Block}) {
		dataWithoutPadding := rx.chunkDecoders[packet.Chunk].Decode()[:packet.Chunk.Size]
		rx.writers[packet.Chunk.FileInfo].WriteAt(dataWithoutPadding, packet.Chunk.Offset)
		rx.finishedChunks[packet.Chunk] = struct{}{}
		delete(rx.chunkDecoders, packet.Chunk) // remove the decoder immediately if the chunk is finished to avoid corruption
	}
}

type Object struct {
	ID   string
	Size int64
}

type Chunk struct {
	FileInfo   Object
	Size       int64
	Offset     int64
	PacketSize int64
}

func (c Chunk) sourceBlockCount() int64 {
	return int64(float64(c.paddedSize() / c.PacketSize))
}
func (c Chunk) decoder() fountain.Decoder {
	return fountain.NewRaptorCodec(int(c.sourceBlockCount()), 8).NewDecoder(int(c.paddedSize()))
}
func (c Chunk) codec() fountain.Codec {
	return fountain.NewRaptorCodec(int(c.sourceBlockCount()), 8)
}
func (c Chunk) targetBlockCount() int64 {
	return c.sourceBlockCount() + 5 // Add a small buffer to allow for Raptor block overflow
}
func (c Chunk) paddedSize() int64 {
	return c.PacketSize * int64(math.Ceil(float64(c.Size)/float64(c.PacketSize)))
}
func (c Chunk) encode(data []byte) []fountain.LTBlock {
	necessaryPadding := c.paddedSize() - c.Size
	paddedData := append(data, make([]byte, necessaryPadding)...)
	return fountain.EncodeLTBlocks(paddedData, c.buildIds(), c.codec())
}
func (c Chunk) buildIds() []int64 {
	ids := make([]int64, c.targetBlockCount())
	for i := range ids {
		ids[i] = int64(i)
	}
	return ids
}

type Packet struct {
	Chunk Chunk
	Block fountain.LTBlock
}
