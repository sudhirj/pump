package pump

import (
	"io"
	"sort"
)

type Transmitter struct {
	readers            map[Object]io.ReaderAt
	chunkPackets       map[Chunk][]Packet
	chunkIndex         int64
	chunkPacketIndexes map[Chunk]int64
}

func NewTransmitter() *Transmitter {
	return &Transmitter{
		readers:            make(map[Object]io.ReaderAt),
		chunkPackets:       make(map[Chunk][]Packet),
		chunkPacketIndexes: make(map[Chunk]int64),
	}
}

func (tx *Transmitter) AddObject(id string, r io.ReaderAt, totalSize int64) (o Object) {
	o.ID = id
	o.Size = totalSize
	tx.readers[o] = r
	return
}
func (tx *Transmitter) ActivateChunkWithWeight(chunk Chunk, weight int) {
	data := make([]byte, chunk.Size)
	tx.readers[chunk.Object].ReadAt(data, chunk.Offset)
	tx.chunkPackets[chunk] = chunk.encode(data)
}
func (tx *Transmitter) GeneratePacket() (packet Packet) {
	chosenChunk := tx.chooseChunk()
	chosenPacketIndex := tx.choosePacketIndex(chosenChunk)
	return tx.chunkPackets[chosenChunk][chosenPacketIndex]
}
func (tx *Transmitter) ActivateChunk(chunk Chunk)   { tx.ActivateChunkWithWeight(chunk, 1) }
func (tx *Transmitter) DeactivateChunk(chunk Chunk) {}

func (tx *Transmitter) chooseChunk() Chunk {
	idx := tx.chunkIndex % int64(len(tx.chunkPackets))
	tx.chunkIndex++
	return tx.activeChunks()[idx]
}
func (tx *Transmitter) activeChunks() (activeChunks []Chunk) {
	for c := range tx.chunkPackets { // Not optimal, but good enough since N is usually small
		activeChunks = append(activeChunks, c)
	}
	sort.Slice(activeChunks, func(i, j int) bool {
		return ((activeChunks[i].Object.ID == activeChunks[j].Object.ID) &&
			(activeChunks[i].Offset < activeChunks[j].Offset)) ||
			(activeChunks[i].Object.ID < activeChunks[j].Object.ID)
	})
	return
}
func (tx *Transmitter) choosePacketIndex(chunk Chunk) int64 {
	idx := tx.chunkPacketIndexes[chunk] % int64(len(tx.chunkPackets[chunk]))
	tx.chunkPacketIndexes[chunk]++
	return idx

}
