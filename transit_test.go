package pump

import (
	"log"
	"math/rand"
	"strconv"
	"testing"
)

func TestLossyTransit(t *testing.T) {
	Size := 4096
	PacketSize := 16

	for lossRate := 0.0; lossRate <= 0.95; lossRate = lossRate + 0.01 {
		virtualFile1 := newVirtualFile(strconv.FormatFloat(lossRate, 'f', 2, 64), int64(Size))

		tx := NewTransmitter()
		sourceFileTxInfo1 := tx.AddObject("s1", virtualFile1, int64(Size))

		tx.ActivateChunk(Chunk{Object: sourceFileTxInfo1, Size: sourceFileTxInfo1.Size, Offset: 0, PacketSize: int64(PacketSize)})
		rx := NewReceiver()
		rx.PrepareForReception(sourceFileTxInfo1, virtualFile1)
		packetCount := 0
		for !rx.Idle() {
			packet := tx.GeneratePacket()
			packetCount++
			if rand.Float64() > lossRate {
				rx.Receive(packet)
			}
		}
		virtualFile1.Validate(t)
		log.Printf("LOSS %.2f / RECOVERY %.2f", lossRate, float32(packetCount)/float32(Size/PacketSize)-1)
	}
}

func TestCompleteSourceLossDuringTransit(t *testing.T) {
	Size := 4096
	PacketSize := 16

	virtualFile1 := newVirtualFile("sourceloss", int64(Size))

	tx := NewTransmitter()
	sourceFileTxInfo1 := tx.AddObject("s1", virtualFile1, int64(Size))

	tx.ActivateChunk(Chunk{Object: sourceFileTxInfo1, Size: sourceFileTxInfo1.Size, Offset: 0, PacketSize: int64(PacketSize)})
	rx := NewReceiver()
	rx.PrepareForReception(sourceFileTxInfo1, virtualFile1)
	packetCount := 0
	for !rx.Idle() {
		packet := tx.GeneratePacket()
		packetCount++
		if packetCount > (Size / PacketSize) {
			rx.Receive(packet)
		}
	}
	virtualFile1.Validate(t)
	log.Printf("RECOVERY EFFICIENCY ORIG %d SENT %d RATIO %.2f", Size/PacketSize, packetCount, float32(packetCount)/float32(Size/PacketSize))

}
