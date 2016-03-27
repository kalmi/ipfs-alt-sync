package ipfs2dir

import (
	"github.com/ipfs/go-ipfs-api"
	"log"
)

func SyncDir(shell *shell.Shell, srcHash string, target string, stat *StatData) {
	list, err := shell.List(srcHash)
	if err != nil {
		log.Fatal(err.Error())
	}
	for _, item := range list {
		task := &SyncTask{
			Shell:     shell,
			Src:       item,
			DstParent: target,
			Stat:      stat,
		}
		RunTask(task)
	}
}
