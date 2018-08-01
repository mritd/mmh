package utils

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
)

const InstallBaseDir = "/usr/bin"

var BinPaths = []string{
	path.Join(InstallBaseDir, "mcp"),
	path.Join(InstallBaseDir, "mec"),
	path.Join(InstallBaseDir, "mgo"),
}

func Install() {

	Uninstall()

	fmt.Println("Install")
	currentPath, err := exec.LookPath(os.Args[0])
	CheckAndExit(err)
	f, err := os.Open(currentPath)
	CheckAndExit(err)
	defer f.Close()
	target, err := os.OpenFile(path.Join(InstallBaseDir, "mmh"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	CheckAndExit(err)
	defer target.Close()

	fmt.Printf("Install %s\n", path.Join(InstallBaseDir, "mmh"))
	_, err = io.Copy(target, f)
	CheckAndExit(err)
	for _, bin := range BinPaths {
		fmt.Printf("Install %s\n", bin)
		err = os.Symlink(path.Join(InstallBaseDir, "mmh"), bin)
		CheckAndExit(err)
	}
}
