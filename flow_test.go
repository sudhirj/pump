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
	sourceFile := generateRandomFile("source", 100)
	sourceFileInfo, _ := sourceFile.Stat()
	defer os.Remove(sourceFile.Name())
	t.Log(sourceFile.Name())

	destinationFile := makeFile("destination")
	defer os.Remove(destinationFile.Name())
	t.Log(destinationFile.Name())

	tx := NewTransmitter()
	sourceFileTxInfo := tx.AddFile("s1", sourceFile, uint64(sourceFileInfo.Size()))
	tx.ActivateChunk(ChunkInfo{FileInfo: sourceFileTxInfo, Size: sourceFileTxInfo.Size, Offset: 0, PacketSize: 10})

	rx := NewReceiver()
	rx.PrepareForReception(sourceFileTxInfo, destinationFile)

	// Run for packet size plus a generous extra
	for i := 0; i <= 10000; i++ {
		rx.Receive(tx.GeneratePacket())
	}
	destinationFile.Sync()
	destinationFileInfo, _ := destinationFile.Stat()

	if sourceFileInfo.Size() != destinationFileInfo.Size() {
		t.Error("Files ought to be same size, but source was", sourceFileInfo.Size(), "and destination was ", destinationFileInfo.Size())
	}

	sourceData, _ := ioutil.ReadFile(sourceFile.Name())
	destinationData, _ := ioutil.ReadFile(destinationFile.Name())
	if !bytes.Equal(sourceData, destinationData) {
		t.Error(sourceData)
		t.Error(destinationData)
		t.Error("File data was not equal")
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
