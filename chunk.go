package pump

import (
	"github.com/google/gofountain"
	"math"
)

type Chunk struct {
	Object     Object
	Size       int64
	Offset     int64
	PacketSize int64
}

type chunkEncoder struct {
	chunk             Chunk
	encoder           fountain.Codec
	sourceSymbols     []fountain.LTBlock
	repairSymbolCache map[int64]fountain.LTBlock
}

func (ce *chunkEncoder) generatePacket(blockIndex int64) Packet {
	if blockIndex <= ce.chunk.sourceBlockCount()-1 {
		return Packet{Chunk: ce.chunk, Block: ce.sourceSymbols[blockIndex]}
	}
	return ce.generateRepairPacket(blockIndex)
}
func (ce *chunkEncoder) data() (data []byte) {
	for _, block := range ce.sourceSymbols {
		data = append(data, block.Data...)
	}
	return
}
func (ce *chunkEncoder) generateRepairPacket(blockIndex int64) Packet {
	if _, availableInCache := ce.repairSymbolCache[blockIndex]; !availableInCache {
		numberOfRepairSymbolsToCache := ce.chunk.sourceBlockCount() / 3
		for _, ltBlock := range fountain.EncodeLTBlocks(ce.data(), buildRange(blockIndex, blockIndex+numberOfRepairSymbolsToCache), ce.encoder) {
			ce.repairSymbolCache[ltBlock.BlockCode] = ltBlock
		}
		// clear old repair symbols out, doesn't make sense to reuse them
		for _, blockCode := range buildRange(blockIndex-numberOfRepairSymbolsToCache, blockIndex-1) {
			delete(ce.repairSymbolCache, blockCode)
		}
	}
	return Packet{Chunk: ce.chunk, Block: ce.repairSymbolCache[blockIndex]}
}

type chunkDecoder struct {
	chunk    Chunk
	decoder  fountain.Decoder
	complete bool
}

func (cd *chunkDecoder) ingest(packet Packet) (finished bool) {
	if cd.complete {
		return // adding blocks to completed decoder will corrupt it
	}
	finished = cd.decoder.AddBlocks([]fountain.LTBlock{packet.Block})
	if finished {
		cd.complete = true
	}
	return
}
func (cd *chunkDecoder) data() []byte {
	return cd.decoder.Decode()[:cd.chunk.Size]
}

func (c Chunk) sourceBlockCount() int64 {
	return int64(float64(c.paddedSize() / c.PacketSize))
}
func (c Chunk) decoder() *chunkDecoder {
	return &chunkDecoder{
		chunk:   c,
		decoder: c.codec().NewDecoder(int(c.paddedSize())),
	}
}
func (c Chunk) codec() fountain.Codec {
	return fountain.NewRaptorCodec(int(c.sourceBlockCount()), 8)
}

func (c Chunk) paddedSize() int64 {
	return c.PacketSize * int64(math.Ceil(float64(c.Size)/float64(c.PacketSize)))
}
func (c Chunk) encode(data []byte) (encoder *chunkEncoder) {
	necessaryPadding := c.paddedSize() - c.Size
	paddedData := append(data, make([]byte, necessaryPadding)...)
	return &chunkEncoder{
		encoder:           c.codec(),
		chunk:             c,
		sourceSymbols:     fountain.EncodeLTBlocks(paddedData, c.buildIds(), c.codec()),
		repairSymbolCache: make(map[int64]fountain.LTBlock),
	}
}
func (c Chunk) buildIds() []int64 {
	return buildRange(0, c.sourceBlockCount())

}
func (c Chunk) valid() bool {
	return c.sourceBlockCount() <= 8000
}
