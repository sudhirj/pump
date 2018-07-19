package pump

import "sort"

type Object struct {
	ID   string
	Size int64
}

func (o Object) isCompletedBy(finishedChunks []Chunk) bool {
	sort.Slice(finishedChunks, func(i, j int) bool {
		return finishedChunks[i].Offset < finishedChunks[j].Offset
	})
	cursor := int64(0)
	for _, chunk := range finishedChunks {
		if o.ID != chunk.Object.ID {
			continue
		}
		if chunk.Offset > cursor {
			return false // next chunk is not adjacent or overlapping
		}
		endOfChunk := chunk.Size + chunk.Offset
		if cursor < endOfChunk { // don't go backward if we already have the data
			cursor = endOfChunk
		}
		if cursor == o.Size { // should have exactly reached the end of the object
			return true
		}
	}
	return false
}
