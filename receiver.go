package pump

import (
	"io"
)

type Receiver struct {
	writers        map[Object]io.WriterAt
	chunkDecoders  map[Chunk]*chunkDecoder
	finishedChunks map[Chunk]struct{}
}

func NewReceiver() *Receiver {
	return &Receiver{
		writers:        make(map[Object]io.WriterAt),
		chunkDecoders:  make(map[Chunk]*chunkDecoder),
		finishedChunks: make(map[Chunk]struct{}),
	}
}

func (rx *Receiver) PrepareForReception(o Object, w io.WriterAt) {
	rx.writers[o] = w
}
func (rx *Receiver) Receive(packet Packet) {
	if _, registered := rx.writers[packet.Chunk.Object]; !registered {
		return
	}
	if _, alreadyFinished := rx.finishedChunks[packet.Chunk]; alreadyFinished {
		return
	}
	if _, present := rx.chunkDecoders[packet.Chunk]; !present {
		rx.chunkDecoders[packet.Chunk] = packet.Chunk.decoder()
	}
	if rx.chunkDecoders[packet.Chunk].Ingest(packet) {
		rx.writers[packet.Chunk.Object].WriteAt(rx.chunkDecoders[packet.Chunk].Data(), packet.Chunk.Offset)
		rx.finishedChunks[packet.Chunk] = struct{}{}
		delete(rx.chunkDecoders, packet.Chunk)
	}
}
