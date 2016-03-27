package main

import (
	"github.com/jawher/mow.cli"
	"github.com/kalmi/ipfs-alt-sync/core/ipfs2dir"
	"github.com/kalmi/ipfs-alt-sync/util"
	"log"
	"os"
)

func main() {
	app := cli.App("ipfs-alt-sync", "Synchronizes files and directories to an unixfs hash. Abuses NTFS alternate streams to store the last syncronization's modification times and hashes for faster dirty change detection.")
	src := app.StringArg("SRC", "", "the source, an ipfs path (/ipfs/... or /ipns/...)")
	dst := app.StringArg("DST", "", "the destination directory (WILL GET OVERWRITTEN!)")
	app.Action = func() {
		shell, err := util.GetLocalShell()
		if err != nil {
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
		var stat ipfs2dir.StatData
		ipfs2dir.SyncDir(shell, sourceHash, *dst, &stat)
		ipfs2dir.PrintStats(&stat)
	}
	app.Run(os.Args)
}
