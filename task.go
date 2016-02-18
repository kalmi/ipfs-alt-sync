package main

import (
	"log"
	"path"
	"sync/atomic"
	"github.com/ipfs/go-ipfs-api"
	"io/ioutil"
	"os"
	"time"
	"math/big"
	"io"
)

const altStreamName = "ipfs-alt-sync"
const maxOutstanding = 4 // Arbitary

var sem = make(chan int, maxOutstanding)

type SyncTask struct {
	shell      *shell.Shell
	dst_parent string
	src        *shell.LsLink
}

func Execute(t *SyncTask) ([]*SyncTask) {
	// Block until there's capacity to process a request
	// (only MaxOutstanding number of requests are allowed to go to the ipfs daemon at the same time)
	sem <- 1
	defer func() {
		<-sem
	}()

	increment(&seenEntityCount)

	switch t.src.Type {
	case shell.TDirectory:
		return processDirectorySyncTask(t)
	case shell.TFile:
		return processFileSyncTask(t)
	default:
		log.Fatal("Unsupported type in unixfs DAG of " + t.src.Name)
		return nil
	}
}

func processDirectorySyncTask(t *SyncTask) ([]*SyncTask) {
	p := path.Join(t.dst_parent, t.src.Name)
	//log.Print("Entering: " + t.src.Hash + " - " + p)
	list, err := t.shell.List(t.src.Hash)
	if err != nil {
		log.Fatal(err.Error())
	}

	sourceDirContentMap := make(dirMap)
	for _, item := range list {
		if item.Type == shell.TDirectory {
			sourceDirContentMap[item.Name] = shell.TDirectory
		} else {
			sourceDirContentMap[item.Name] = shell.TFile
		}
	}

	targetDirContent, err := ioutil.ReadDir(p)
	for _, item := range targetDirContent {
		fileType := shell.TRaw
		if item.IsDir() {
			fileType = shell.TDirectory
		} else {
			fileType = shell.TFile
		}

		if fileType != sourceDirContentMap[item.Name()] {
			//log.Print("Removing: " + path.Join(p, item.Name()))
			if fileType == shell.TDirectory {
				err := os.RemoveAll(path.Join(p, item.Name()))
				if err != nil {
					log.Fatal(err.Error())
				}
			} else {
				err := os.Remove(path.Join(p, item.Name()))
				if err != nil {
					log.Fatal(err.Error())
				}
			}
		}
	}

	tasks := make([]*SyncTask, len(list), len(list))
	for i, item := range list {
		tasks[i] = &SyncTask{
			shell: t.shell,
			dst_parent: p,
			src: item,
		}
	}
	return tasks
}

func processFileSyncTask(t *SyncTask) ([]*SyncTask) {
	increment(&seenFileCount)
	p := path.Join(t.dst_parent, t.src.Name)
	//log.Print("Looking at: " + t.src.Hash + " - " + p)
	if fileQuickMatches(p, t.src.Hash) {
		//log.Print("Skipped: " + t.src.Hash + " - " + p)
		increment(&skippedEntityCount)
		return nil
	} else {
		r, err := t.shell.Cat(t.src.Hash)
		if err != nil {
			log.Fatal(err.Error())
		}

		outFile, err := os.Create(p)
		if err != nil {
			// Try creating the container directory
			err := os.MkdirAll(t.dst_parent, os.ModePerm) //TODO: race condition?
			if err != nil {
				// TODO: Is ENOTDIR possible here? If it is, then that should be handled by deleting the non-dir path. Maybe it is not possible if we process the entries in a sorted manner.
				log.Fatal(err.Error())
			}
			//Okay, now lets try again with the parent directory in place now.
			outFile, err = os.Create(p)
			if err != nil {
				log.Fatal(err.Error())
			}
		}

		_, err = io.Copy(outFile, r)
		if err != nil {
			log.Fatal(err.Error())
		}

		outFile.Close()

		stat, err := os.Stat(p)
		if err != nil {
			log.Fatal(err.Error())
		}

		alternateStream, err := os.Create(p + ":" + altStreamName)
		if err != nil {
			log.Fatal(err.Error())
		}

		c := altStreamContentFormatter(t.src.Hash, stat.ModTime())
		_, err = alternateStream.WriteString(c)
		if err != nil {
			log.Fatal(err.Error())
		}
		err = alternateStream.Close()
		if err != nil {
			log.Fatal(err.Error())
		}

		//log.Print("Done: " + t.src.Hash + " - " + p)
		increment(&processedEntityCount)
		return nil // It was not a directory, so no new tasks were created
	}
}
func fileQuickMatches(path string, hash string) bool {
	bytes, err := ioutil.ReadFile(path + ":" + altStreamName)
	if err == nil {
		stat, err := os.Stat(path)
		if err == nil {
			if string(bytes) == altStreamContentFormatter(hash, stat.ModTime()) {
				return true
			}
		}
	}
	return false
}

func altStreamContentFormatter(hash string, modTime time.Time) string {
	buf := big.NewInt(modTime.Unix()).String()
	return hash + "|" + string(buf)
}

func increment(i *uint64) {
	atomic.AddUint64(i, 1)
}