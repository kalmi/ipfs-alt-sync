package main

import (
	"log"
	"github.com/jawher/mow.cli"
	"os"
	"github.com/ipfs/go-ipfs-api"
	"path"
	"io"
	"io/ioutil"
	"math/big"
)

var (
	src *string
	dst *string
	processedFileCount int
	skippedFileCount int
)

type DirMap map[string]int

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

	sync(shell, sourceHash, *dst)

	log.Print("Processed file count:")
	log.Print(processedFileCount)
	log.Print("Skipped (already up-to-date) file count:")
	log.Print(skippedFileCount)
}




func sync(s *shell.Shell, srcHash string, target string) {
	list, err := s.List(srcHash)
	if err != nil {
		log.Fatal(err.Error())
	}

	sourceDirContentMap := make(DirMap)
	for _, item := range list {
		if item.Type == shell.TDirectory {
			sourceDirContentMap[item.Name] = shell.TDirectory
		} else {
			sourceDirContentMap[item.Name] = shell.TFile
		}
	}

	targetDirContent, err := ioutil.ReadDir(target)
	for _, item := range targetDirContent {
		fileType := shell.TRaw
		if item.IsDir(){
			fileType = shell.TDirectory
		} else {
			fileType = shell.TFile
		}

		if fileType != sourceDirContentMap[item.Name()] {
			log.Print("Removing " + path.Join(target, item.Name()))
			if fileType == shell.TDirectory {
				err := os.RemoveAll(path.Join(target, item.Name()))
				if err != nil {
					log.Fatal(err.Error())
				}
			} else {
				err := os.Remove(path.Join(target, item.Name()))
				if err != nil {
					log.Fatal(err.Error())
				}
			}
		}

	}


	for _, item := range list {
		log.Print(item.Hash + " - " + item.Name + " - " + target)
		switch item.Type {
		case shell.TDirectory:
			sync(s, item.Hash, path.Join(target, item.Name))

		case shell.TFile:
			processedFileCount += 1

			itemPathOnFs := path.Join(target, item.Name)
			bytes, err := ioutil.ReadFile(itemPathOnFs + ":ipfs-alt-sync")
			if err == nil {
				stat, err := os.Stat(itemPathOnFs)
				if err == nil {
					buf := big.NewInt(stat.ModTime().Unix()).String()
					if string(bytes) == (item.Hash + "|" + string(buf)) {
						skippedFileCount += 1
						continue
					}
				}
			}

			r, err := s.Cat(item.Hash)
			if err != nil {
				log.Fatal(err.Error())
			}

			outFile, err := os.Create(itemPathOnFs)
			if err != nil {
				// Try creating the container directory
				// (ipfs generates unixfs structures that don't contain directories for non-empty directories)
				err := os.MkdirAll(target, os.ModePerm)
				if err != nil {
					// TODO: Is ENOTDIR possible here? If it is, then that should be handled by deleting the non-dir path. Maybe it is not possible if we process the entries in a sorted manner.
					log.Fatal(err.Error())
				}
				//Okay, now lets try again with the parent directory in place now.
				outFile, err = os.Create(itemPathOnFs)
				if err != nil {
					log.Fatal(err.Error())
				}
			}

			_, err = io.Copy(outFile, r)
			if err != nil {
				log.Fatal(err.Error())
			}

			outFile.Close()

			stat, err := os.Stat(itemPathOnFs)
			if err != nil {
				log.Fatal(err.Error())
			}

			alternateStream, err := os.Create(itemPathOnFs + ":ipfs-alt-sync")
			if err != nil {
				log.Fatal(err.Error())
			}

			buf := big.NewInt(stat.ModTime().Unix()).String()
			alternateStream.WriteString(item.Hash + "|" + string(buf))
			alternateStream.Close()

		default:
			log.Fatal("Unsupported item type present in ipfs stucture.")
		}
	}
}
