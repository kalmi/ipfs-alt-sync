package util

import "log"
import "sync/atomic"

type StatData struct {
	seenEntityCount uint64
	seenFileCount uint64
	processedEntityCount uint64
	skippedEntityCount uint64
}

func incrementSeenEntityCount(stat *StatData) {
	atomic.AddUint64(&stat.seenEntityCount, 1)
}

func incrementSeenFileCount(stat *StatData) {
	atomic.AddUint64(&stat.seenFileCount, 1)
}

func incrementProcessedEntityCount(stat *StatData) {
	atomic.AddUint64(&stat.processedEntityCount, 1)
}

func incrementSkippedEntityCount(stat *StatData) {
	atomic.AddUint64(&stat.skippedEntityCount, 1)
}

func PrintStats(stat *StatData) {
	log.Print("Seen entity count:")
	log.Print(atomic.LoadUint64(&stat.seenEntityCount))
	log.Print("Seen file count:")
	log.Print(atomic.LoadUint64(&stat.seenFileCount))
	log.Print("Overwritten file count:")
	log.Print(atomic.LoadUint64(&stat.processedEntityCount))
	log.Print("Skipped (already up-to-date) file count:")
	log.Print(atomic.LoadUint64(&stat.skippedEntityCount))
}