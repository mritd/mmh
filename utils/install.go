package utils

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

func Install(dir string) {

	var binPaths = []string{
		filepath.Join(dir, "mcp"),
		filepath.Join(dir, "mec"),
		filepath.Join(dir, "mgo"),
		filepath.Join(dir, "mcs"),
		filepath.Join(dir, "mcx"),
		filepath.Join(dir, "mping"),
	}

	currentPath, err := exec.LookPath(os.Args[0])
	CheckAndExit(err)

	if !Root() {
		cmd := exec.Command("sudo", currentPath, "install", "--dir", dir)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		CheckAndExit(cmd.Run())
	} else {

		Uninstall(dir)

		f, err := os.Open(currentPath)
		CheckAndExit(err)
		defer func() { _ = f.Close() }()

		target, err := os.OpenFile(filepath.Join(dir, "mmh"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
		CheckAndExit(err)
		defer func() { _ = target.Close() }()

		fmt.Printf("📥 install %s\n", filepath.Join(dir, "mmh"))
		_, err = io.Copy(target, f)
		CheckAndExit(err)
		for _, bin := range binPaths {
			fmt.Printf("📥 install %s\n", bin)
			CheckAndExit(os.Symlink(filepath.Join(dir, "mmh"), bin))
		}
	}

}
