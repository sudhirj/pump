package pump

import (
	"github.com/google/gofountain"
	"math"
)

type FountainBlock = fountain.LTBlock

type Chunk struct {
	Object     Object
	Size       int64
	Offset     int64
	PacketSize int64
}

type chunkEncoder struct {
	chunk       Chunk
	encoder     fountain.Codec
	data        []byte
	symbolCache map[int64]FountainBlock
}

func (ce *chunkEncoder) generatePacket(blockIndex int64) Packet {
	if _, available := ce.symbolCache[blockIndex]; !available {
		idsToBuild := buildRange(blockIndex, blockIndex+ce.chunk.sourceBlockCount())
		blocks := fountain.EncodeLTBlocks(ce.copyOfData(), idsToBuild, ce.encoder)
		ce.symbolCache = make(map[int64]FountainBlock)
		for _, block := range blocks {
			ce.symbolCache[block.BlockCode] = block
		}
	}
	return Packet{Chunk: ce.chunk, Block: ce.symbolCache[blockIndex]}
}
func (ce *chunkEncoder) copyOfData() []byte {
	dataCopy := make([]byte, len(ce.data))
	copy(dataCopy, ce.data)
	return dataCopy
}

type chunkDecoder struct {
	chunk    Chunk
	decoder  fountain.Decoder
	complete bool
}

func (cd *chunkDecoder) ingest(packet Packet) (finished bool) {
	if cd.complete {
		return // because adding blocks to completed decoder will corrupt it
	}
	finished = cd.decoder.AddBlocks([]FountainBlock{packet.Block})
	if finished {
		cd.complete = true
	}
	return
}
func (cd *chunkDecoder) data() []byte {
	return cd.decoder.Decode()[:cd.chunk.Size]
}

func (c Chunk) sourceBlockCount() int64 {
	return int64(math.Ceil(float64(c.alignedSourceBlockSize()) / float64(c.PacketSize)))
}
func (c Chunk) decoder() *chunkDecoder {
	return &chunkDecoder{
		chunk:   c,
		decoder: c.codec().NewDecoder(int(c.alignedSourceBlockSize())),
	}
}
func (c Chunk) codec() fountain.Codec {
	return fountain.NewRaptorCodec(int(c.sourceBlockCount()), 8)
}
func (c Chunk) valid() bool {
	return c.sourceBlockCount() <= 8100
}
func (c Chunk) encoder(data []byte) (encoder *chunkEncoder) {
	return &chunkEncoder{
		encoder:     c.codec(),
		chunk:       c,
		data:        append(data, c.padding()...),
		symbolCache: make(map[int64]FountainBlock),
	}
}
func (c Chunk) alignedSourceBlockSize() int64 {
	return c.PacketSize * int64(math.Ceil(float64(c.Size)/float64(c.PacketSize)))
}

func (c Chunk) buildIds() []int64 {
	return buildRange(0, c.sourceBlockCount()*3)

}

func (c Chunk) padding() []byte {
	return make([]byte, c.alignedSourceBlockSize()-c.Size)
}
