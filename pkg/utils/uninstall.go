package utils

import (
	"fmt"
	"os"
	"path"
)

func Uninstall() {
	CheckRoot()

	fmt.Println("Uninstall")

	for _, bin := range BinPaths {
		fmt.Printf("Remove %s\n", bin)
		os.Remove(bin)
	}
	fmt.Printf("Remove %s\n", path.Join(InstallBaseDir, "mmh"))
	os.Remove(path.Join(InstallBaseDir, "mmh"))
}
