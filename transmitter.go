package pump

import (
	"errors"
	"io"
)

var ChunkRatioInvalid = errors.New("the ratio of chunk to packet size is too high - either use larger packets or a smaller chunk")

type Transmitter struct {
	readers            map[Object]io.ReaderAt
	chunkEncoders      map[Chunk]*chunkEncoder
	chunkIndex         int64
	chunkPacketIndexes map[Chunk]int64
}

func NewTransmitter() *Transmitter {
	return &Transmitter{
		readers:            make(map[Object]io.ReaderAt),
		chunkEncoders:      make(map[Chunk]*chunkEncoder),
		chunkPacketIndexes: make(map[Chunk]int64),
	}
}

func (tx *Transmitter) AddObject(id string, r io.ReaderAt, totalSize int64) (o Object) {
	o.ID = id
	o.Size = totalSize
	tx.readers[o] = r
	return
}
func (tx *Transmitter) ActivateChunk(chunk Chunk) error {
	if !chunk.valid() {
		return ChunkRatioInvalid
	}
	data := make([]byte, chunk.Size)
	tx.readers[chunk.Object].ReadAt(data, chunk.Offset)
	tx.chunkEncoders[chunk] = chunk.encoder(data)
	return nil
}
func (tx *Transmitter) GeneratePacket() (packet Packet) {
	chosenChunk := tx.chooseChunk()
	chosenPacketIndex := tx.choosePacketIndex(chosenChunk)
	return tx.chunkEncoders[chosenChunk].generatePacket(chosenPacketIndex)
}
func (tx *Transmitter) DeactivateChunk(chunk Chunk) {}

func (tx *Transmitter) chooseChunk() Chunk {
	idx := tx.chunkIndex % int64(len(tx.chunkEncoders))
	tx.chunkIndex++
	return tx.activeChunks()[idx]
}
func (tx *Transmitter) activeChunks() (activeChunks []Chunk) {
	for c := range tx.chunkEncoders { // Not optimal, but good enough since N is usually small
		activeChunks = append(activeChunks, c)
	}
	return
}
func (tx *Transmitter) choosePacketIndex(chunk Chunk) int64 {
	idx := tx.chunkPacketIndexes[chunk]
	tx.chunkPacketIndexes[chunk]++
	return idx

}
