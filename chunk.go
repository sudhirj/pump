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
	chunk       Chunk
	encoder     fountain.Codec
	data        []byte
	symbolCache map[int64]fountain.LTBlock
}

func (ce *chunkEncoder) generatePacket(blockIndex int64) Packet {
	if _, available := ce.symbolCache[blockIndex]; !available {
		tempData := make([]byte, len(ce.data))
		copy(tempData, ce.data)
		blocks := fountain.EncodeLTBlocks(tempData, buildRange(blockIndex, blockIndex+ce.chunk.sourceBlockCount()), ce.encoder)
		ce.symbolCache = make(map[int64]fountain.LTBlock)
		for _, block := range blocks {
			ce.symbolCache[block.BlockCode] = block
		}
	}
	return Packet{Chunk: ce.chunk, Block: ce.symbolCache[blockIndex]}
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
	return int64(math.Ceil(float64(c.paddedSourceBlockSize()) / float64(c.PacketSize)))
}
func (c Chunk) decoder() *chunkDecoder {
	return &chunkDecoder{
		chunk:   c,
		decoder: c.codec().NewDecoder(int(c.paddedSourceBlockSize())),
	}
}
func (c Chunk) codec() fountain.Codec {
	return fountain.NewRaptorCodec(int(c.sourceBlockCount()), 8)
}
func (c Chunk) valid() bool {
	return c.sourceBlockCount() <= 8100
}
func (c Chunk) encoder(data []byte) (encoder *chunkEncoder) {
	necessaryPadding := c.paddedSourceBlockSize() - c.Size
	paddedData := append(data, make([]byte, necessaryPadding)...)
	return &chunkEncoder{
		encoder:     c.codec(),
		chunk:       c,
		data:        paddedData,
		symbolCache: make(map[int64]fountain.LTBlock),
	}
}
func (c Chunk) paddedSourceBlockSize() int64 {
	return c.PacketSize * int64(math.Ceil(float64(c.Size)/float64(c.PacketSize)))
}

func (c Chunk) buildIds() []int64 {
	return buildRange(0, c.sourceBlockCount()*3)

}
