package ipfs2dir

import "log"
import "sync/atomic"

type statDataType uint64

const (
	SeenEntityCount      statDataType = iota
	SeenFileCount                     = iota
	ProcessedEntityCount              = iota
	SkippedEntityCount                = iota
	// The next line is needed for the array size in the StatData type
	numberOfDataPoints = iota
)

type StatData struct {
	data [numberOfDataPoints]uint64
}

func (s *StatData) Increment(whatToIncrement statDataType) {
	atomic.AddUint64(&s.data[whatToIncrement], 1)
}

func (s *StatData) Read(whatToRead statDataType) uint64 {
	return atomic.LoadUint64(&s.data[whatToRead])
}

func (s *StatData) Print() {
	log.Print("Seen entity count:")
	log.Print(s.Read(SeenEntityCount))
	log.Print("Seen file count:")
	log.Print(s.Read(SeenFileCount))
	log.Print("Overwritten file count:")
	log.Print(s.Read(ProcessedEntityCount))
	log.Print("Skipped (already up-to-date) file count:")
	log.Print(s.Read(SkippedEntityCount))
}
