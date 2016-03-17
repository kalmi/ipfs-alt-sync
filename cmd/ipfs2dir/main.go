package main

import (
	"os"
	"log"
	util "github.com/kalmi/ipfs-alt-sync/util"
	ipfs2dir "github.com/kalmi/ipfs-alt-sync/cmd/ipfs2dir/util"
	"github.com/jawher/mow.cli"
	"github.com/ipfs/go-ipfs-api"
)

var (
	src *string
	dst *string
	stat ipfs2dir.StatData
)

func main() {

	app := cli.App("ipfs-alt-sync", "Synchronizes files and directories to an unixfs hash. Abuses NTFS alternate streams to store the last syncronization's modification times and hashes for faster dirty change detection.")
	app.Action = action
	src = app.StringArg("SRC", "", "the source, an ipfs path (/ipfs/... or /ipns/...)")
	dst = app.StringArg("DST", "", "the destination directory (WILL GET OVERWRITTEN!)")
	app.Run(os.Args)
}

func action() {
	shell, err := util.GetLocalShell()
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
	ipfs2dir.PrintStats(&stat)
}

func syncDir(shell *shell.Shell, srcHash string, target string) {
	list, err := shell.List(srcHash)
	if err != nil {
		log.Fatal(err.Error())
	}
	for _, item := range(list) {
		task := &ipfs2dir.SyncTask{
			Shell: shell,
			Src: item,
			DstParent: target,
			Stat: &stat,
		}
		ipfs2dir.RunTask(task)
	}
}

