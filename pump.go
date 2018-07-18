package pump

import (
	"github.com/google/gofountain"
	"log"
	"time"
)

type Packet struct {
	Chunk Chunk
	Block fountain.LTBlock
}

func buildRange(start int64, end int64) []int64 {
	ids := make([]int64, end-start)
	for i := range ids {
		ids[i] = start + int64(i)
	}
	return ids
}

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}
