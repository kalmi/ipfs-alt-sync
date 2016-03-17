package util

import (
	"os"
	"io"
	"io/ioutil"
	"log"
	"path"
	"sync"
	"github.com/ipfs/go-ipfs-api"
	"github.com/kalmi/ipfs-alt-sync/util"
)

const maxOutstanding = 4 // Arbitary

var sem = make(chan int, maxOutstanding)
var lock sync.Mutex

type dirMap map[string]int

type SyncTask struct {
	Shell     *shell.Shell
	DstParent string
	Src       *shell.LsLink
	Stat      *StatData
}

func Execute(t *SyncTask) ([]*SyncTask) {
	// Block until there's capacity to process a request
	// (only MaxOutstanding number of requests are allowed to go to the ipfs daemon at the same time)
	sem <- 1
	defer func() {
		<-sem
	}()

	incrementSeenEntityCount(t.Stat)

	switch t.Src.Type {
	case shell.TDirectory:
		return processDirectorySyncTask(t)
	case shell.TFile:
		return processFileSyncTask(t)
	default:
		log.Fatal("Unsupported type in unixfs DAG of " + t.Src.Name)
		return nil
	}
}

func processDirectorySyncTask(t *SyncTask) ([]*SyncTask) {
	p := path.Join(t.DstParent, t.Src.Name)
	//log.Print("Entering: " + t.src.Hash + " - " + p)
	list, err := t.Shell.List(t.Src.Hash)
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
			Shell: t.Shell,
			DstParent: p,
			Src: item,
			Stat: t.Stat,
		}
	}
	return tasks
}

func processFileSyncTask(t *SyncTask) ([]*SyncTask) {
	incrementSeenFileCount(t.Stat)
	p := path.Join(t.DstParent, t.Src.Name)
	//log.Print("Looking at: " + t.src.Hash + " - " + p)
	if util.FileQuickMatches(p, t.Src.Hash) {
		//log.Print("Skipped: " + t.src.Hash + " - " + p)
		incrementSkippedEntityCount(t.Stat)
		return nil
	} else {
		r, err := t.Shell.Cat(t.Src.Hash)
		if err != nil {
			log.Fatal(err.Error())
		}

		outFile, err := os.Create(p)
		if err != nil {
			// Try creating the container directory
			lock.Lock()
			err := os.MkdirAll(t.DstParent, os.ModePerm)
			lock.Unlock()
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

		util.TagFile(p, util.TagData{
			Hash: t.Src.Hash,
			ModTime: stat.ModTime(),
		})

		//log.Print("Done: " + t.src.Hash + " - " + p)
		incrementProcessedEntityCount(t.Stat)
		return nil // It was not a directory, so no new tasks were created
	}
}

func RunTask(task *SyncTask){
	tasks := Execute(task)

	if len(tasks) > 0 {
		var wg sync.WaitGroup
		wg.Add(len(tasks))
		for _, task := range tasks {
			go func(task *SyncTask){
				defer wg.Done()
				RunTask(task)
			}(task)
		}
		wg.Wait()
	}
}
