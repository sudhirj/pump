package pump

import (
	"github.com/google/gofountain"
)

type Packet struct {
	Chunk Chunk
	Block fountain.LTBlock
}
