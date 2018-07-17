package pump

import (
	"github.com/google/gofountain"
	"math"
)

type Object struct {
	ID   string
	Size int64
}

type Chunk struct {
	Object     Object
	Size       int64
	Offset     int64
	PacketSize int64
}

type ChunkDecoder struct {
	Chunk    Chunk
	decoder  fountain.Decoder
	complete bool
}

func (cd *ChunkDecoder) Ingest(packet Packet) (finished bool) {
	if cd.complete {
		return
	}
	finished = cd.decoder.AddBlocks([]fountain.LTBlock{packet.Block})
	if finished {
		cd.complete = true
	}
	return
}
func (cd *ChunkDecoder) Data() []byte {
	return cd.decoder.Decode()[:cd.Chunk.Size]
}

func (c Chunk) sourceBlockCount() int64 {
	return int64(float64(c.paddedSize() / c.PacketSize))
}
func (c Chunk) decoder() *ChunkDecoder {
	return &ChunkDecoder{
		Chunk:   c,
		decoder: c.codec().NewDecoder(int(c.paddedSize())),
	}
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
func (c Chunk) encode(data []byte) (packets []Packet) {
	necessaryPadding := c.paddedSize() - c.Size
	paddedData := append(data, make([]byte, necessaryPadding)...)
	for _, ltBlock := range fountain.EncodeLTBlocks(paddedData, c.buildIds(), c.codec()) {
		packets = append(packets, Packet{Chunk: c, Block: ltBlock})
	}
	return
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
