package util

import (
	"log"
	"time"
	"math/big"
	"io/ioutil"
	"os"
)

const altStreamName = "ipfs-alt-sync"

type TagData struct {
	Hash string
	ModTime time.Time
}

func AltStreamContentFormatter(tagData TagData) string {
	buf := big.NewInt(tagData.ModTime.Unix()).String()
	return tagData.Hash + "|" + string(buf)
}

func FileQuickMatches(path string, hash string) bool {
	bytes, err := ioutil.ReadFile(path + ":" + altStreamName)
	if err == nil {
		stat, err := os.Stat(path)
		if err == nil {
			var tagData = TagData{
				Hash: hash,
				ModTime: stat.ModTime(),
			}
			if string(bytes) == AltStreamContentFormatter(tagData) {
				return true
			}
		}
	}
	return false
}

func TagFile(path string, tagData TagData) {
	//TODO: better error handling

	alternateStream, err := os.Create(path + ":" + altStreamName)
	if err != nil {
		log.Fatal(err.Error())
	}

	c := AltStreamContentFormatter(tagData)
	_, err = alternateStream.WriteString(c)
	if err != nil {
		log.Fatal(err.Error())
	}
	err = alternateStream.Close()
	if err != nil {
		log.Fatal(err.Error())
	}
}