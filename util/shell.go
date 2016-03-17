package util

import (
	"os"
	"strings"
	"io/ioutil"
	"path/filepath"

	hd "github.com/mitchellh/go-homedir"
	sh "github.com/ipfs/go-ipfs-api"
	manet "github.com/jbenet/go-multiaddr-net"
	ma "github.com/jbenet/go-multiaddr-net/Godeps/_workspace/src/github.com/jbenet/go-multiaddr"
)

func GetLocalShell() (*sh.Shell, error) {
	apiAddress := os.Getenv("IPFS_API")
	if apiAddress != "" {
		// IPFS_API takes priority.
		return sh.NewShell(apiAddress), nil
	}

	apiFilePath := ""
	ipath := os.Getenv("IPFS_PATH")
	if ipath != "" {
		apiFilePath = filepath.Join(ipath, "api")
	}
	home, err := hd.Dir() // Get home directory
	if err == nil {
		apiFilePath = filepath.Join(home, ".ipfs", "api")
	}

	// Read the file (hopefully) containing the location of an ipfs daemon
	apiFileContent, err := ioutil.ReadFile(apiFilePath)
	if err != nil {
		return nil, err
	}

	multiAddr := strings.Trim(string(apiFileContent), "\n\t ")
	apiAddress, err = multiaddrToNormal(multiAddr)
	if err != nil {
		return nil, err
	}
	return sh.NewShell(apiAddress), nil
}

func multiaddrToNormal(addr string) (string, error) {
	// Taken from gx's source

	maddr, err := ma.NewMultiaddr(addr)
	if err != nil {
		return "", err
	}

	_, host, err := manet.DialArgs(maddr)
	if err != nil {
		return "", err
	}

	return host, nil
}