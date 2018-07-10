package chunkasaur

import (
	"testing"
	"io/ioutil"
	"math/rand"
	"time"
	"os"
	"bytes"
)

func TestSingleChunkTransmission(t *testing.T) {
	Size := 100000
	PacketSize := 100
	SymbolCount := Size / PacketSize
	TransmissionBuffer := 10

	sourceFile := generateRandomFile("source", Size)
	defer os.Remove(sourceFile.Name())

	destinationFile := makeFile("destination")
	defer os.Remove(destinationFile.Name())

	tx := NewTransmitter()
	sourceFileTxInfo := tx.AddFile("s1", sourceFile, int64(Size))
	tx.ActivateChunk(Chunk{FileInfo: sourceFileTxInfo, Size: sourceFileTxInfo.Size, Offset: 0, PacketSize: int64(PacketSize)})

	rx := NewReceiver()
	rx.PrepareForReception(sourceFileTxInfo, destinationFile)

	// Run for symbol count plus a buffer

	for i := 0; i <= SymbolCount+TransmissionBuffer; i++ {
		rx.Receive(tx.GeneratePacket())
	}

	assertEquality(t, sourceFile, destinationFile)
}

func assertEquality(t *testing.T, sourceFile, destinationFile *os.File) {
	sourceFile.Sync()
	destinationFile.Sync()
	sourceFileInfo, _ := destinationFile.Stat()
	destinationFileInfo, _ := destinationFile.Stat()
	if sourceFileInfo.Size() != destinationFileInfo.Size() {
		t.Error("Files ought to be same size, but source was", sourceFileInfo.Size(), "and destination was ", destinationFileInfo.Size())
	}

	sourceData, _ := ioutil.ReadFile(sourceFile.Name())
	destinationData, _ := ioutil.ReadFile(destinationFile.Name())
	if !bytes.Equal(sourceData, destinationData) {
		diffCount := 0
		for _, i := range sourceData {
			if sourceData[i] != destinationData[i] {
				diffCount++
			}
		}
		t.Error("File data was not equal, diffcount", diffCount)
	}
}

func generateRandomFile(prefix string, size int) (*os.File) {
	sourceFile := makeFile(prefix)
	sourceFile.Write(randomBytes(size))
	sourceFile.Sync()
	return sourceFile
}

func makeFile(prefix string) (*os.File) {
	sourceFile, _ := ioutil.TempFile("", prefix)
	return sourceFile
}
func randomBytes(num int) []byte {
	rand.Seed(time.Now().UnixNano())
	holder := make([]byte, num)
	rand.Read(holder)
	return holder
}
