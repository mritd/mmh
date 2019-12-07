package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func Uninstall(dir string) {

	var binPaths = []string{
		filepath.Join(dir, "mcp"),
		filepath.Join(dir, "mec"),
		filepath.Join(dir, "mgo"),
		filepath.Join(dir, "mcs"),
		filepath.Join(dir, "mcf"),
		filepath.Join(dir, "mping"),
		filepath.Join(dir, "mtun"),
	}

	currentPath, err := exec.LookPath(os.Args[0])
	CheckAndExit(err)

	if !Root() {
		cmds := append(os.Environ(), currentPath, "uninstall", "--dir", dir)
		cmd := exec.Command("sudo", cmds...)
		cmd.Env = os.Environ()
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		CheckAndExit(cmd.Run())
	} else {
		for _, bin := range binPaths {
			fmt.Printf("👉 remove %s\n", bin)
			_ = os.Remove(bin)
		}
		fmt.Printf("👉 remove %s\n", filepath.Join(dir, "mmh"))
		_ = os.Remove(filepath.Join(dir, "mmh"))
	}
}
