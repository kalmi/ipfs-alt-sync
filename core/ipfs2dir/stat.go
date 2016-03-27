package ipfs2dir

import "log"
import "sync/atomic"

type statDataType uint64

const (
	SeenEntityCount      statDataType = iota
	SeenFileCount        statDataType = iota
	ProcessedEntityCount statDataType = iota
	SkippedEntityCount   statDataType = iota
	// The next line is needed for the array size in the StatData type
	numberOfDataPoints statDataType = iota
)

type StatData struct {
	data [numberOfDataPoints]uint64
}

func (s *StatData) Increment(whatToIncrement statDataType) {
	atomic.AddUint64(&s.data[whatToIncrement], 1)
}

func (s *StatData) Print() {
	log.Print("Seen entity count:")
	log.Print(atomic.LoadUint64(&s.data[SeenEntityCount]))
	log.Print("Seen file count:")
	log.Print(atomic.LoadUint64(&s.data[SeenFileCount]))
	log.Print("Overwritten file count:")
	log.Print(atomic.LoadUint64(&s.data[ProcessedEntityCount]))
	log.Print("Skipped (already up-to-date) file count:")
	log.Print(atomic.LoadUint64(&s.data[SkippedEntityCount]))
}
