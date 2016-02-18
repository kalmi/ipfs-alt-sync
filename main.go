package main

import (
	"log"
	"github.com/jawher/mow.cli"
	"os"
	"github.com/ipfs/go-ipfs-api"
	"sync"
	"sync/atomic"
)

var (
	src *string
	dst *string
	seenEntityCount uint64
	seenFileCount uint64
	processedEntityCount uint64
	skippedEntityCount uint64
)

type dirMap map[string]int

func main() {

	app := cli.App("ipfs-alt-sync", "Synchronizes files and directories to an unixfs hash. Abuses NTFS alternate streams to store the last syncronization's modification times and hashes for faster dirty change detection.")
	app.Action = action
	src = app.StringArg("SRC", "", "the source, an ipfs path (/ipfs/... or /ipns/...)")
	dst = app.StringArg("DST", "", "the destination directory (WILL GET OVERWRITTEN!)")
	app.Run(os.Args)
}

func action() {
	shell, err := getLocalShell()
	if (err != nil) {
		log.Fatal("Could not find ipfs daemon: " + err.Error())
	}

	if !shell.IsUp() {
		log.Fatal("Daemon is not up")
	}

	sourceHash, err := shell.ResolvePath(*src)
	if err != nil {
		log.Fatal("Specified source could not be resolved: " + err.Error())
	}

	log.Print("Source hash resolved to " + sourceHash)

	destinationDir, err := os.Stat(*dst)
	if err != nil && os.IsNotExist(err) {
		log.Fatal("Destination directory does not exist.")
	} else if err != nil {
		log.Fatal("Something is wrong with the destination directory: " + err.Error())
	} else if !destinationDir.IsDir() {
		log.Fatal("Destination is not a directory.")
	}

	syncDir(shell, sourceHash, *dst)

	log.Print("Seen entity count:")
	log.Print(atomic.LoadUint64(&seenEntityCount))
	log.Print("Seen file count:")
	log.Print(atomic.LoadUint64(&seenFileCount))
	log.Print("Overwritten file count:")
	log.Print(atomic.LoadUint64(&processedEntityCount))
	log.Print("Skipped (already up-to-date) file count:")
	log.Print(atomic.LoadUint64(&skippedEntityCount))
}

func syncDir(shell *shell.Shell, srcHash string, target string) {
	list, err := shell.List(srcHash)
	if err != nil {
		log.Fatal(err.Error())
	}
	for _, item := range(list) {
		task := &SyncTask{
			shell: shell,
			src: item,
			dst_parent: target,
		}
		runTask(task)
	}
}

func runTask(task *SyncTask){
	tasks := Execute(task)

	if len(tasks) > 0 {
		var wg sync.WaitGroup
		wg.Add(len(tasks))
		for _, task := range tasks {
			go func(task *SyncTask){
				defer wg.Done()
				runTask(task)
			}(task)
		}
		wg.Wait()
	}
}

