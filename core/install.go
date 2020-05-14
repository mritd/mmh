package core

import (
	"fmt"
	"io"
	"os"
	osexec "os/exec"
	"path/filepath"
)

func Install(dir string) {
	var binPaths []string
	for _, as := range Aliases {
		binPaths = append(binPaths, filepath.Join(dir, as))
	}

	currentPath, err := osexec.LookPath(os.Args[0])
	checkAndExit(err)

	if !isRoot() {
		cmds := append(os.Environ(), currentPath, "install", "--dir", dir)
		cmd := osexec.Command("sudo", cmds...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		checkAndExit(cmd.Run())
	} else {
		Uninstall(dir)
		f, err := os.Open(currentPath)
		checkAndExit(err)
		defer func() { _ = f.Close() }()

		target, err := os.OpenFile(filepath.Join(dir, "mmh"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
		checkAndExit(err)
		defer func() { _ = target.Close() }()

		fmt.Printf("📥 install %s\n", filepath.Join(dir, "mmh"))
		_, err = io.Copy(target, f)
		checkAndExit(err)
		for _, bin := range binPaths {
			fmt.Printf("📥 install %s\n", bin)
			checkAndExit(os.Symlink(filepath.Join(dir, "mmh"), bin))
		}
	}

}

func Uninstall(dir string) {
	var binPaths []string
	for _, as := range Aliases {
		binPaths = append(binPaths, filepath.Join(dir, as))
	}

	currentPath, err := osexec.LookPath(os.Args[0])
	checkAndExit(err)

	if !isRoot() {
		cmds := append(os.Environ(), currentPath, "uninstall", "--dir", dir)
		cmd := osexec.Command("sudo", cmds...)
		cmd.Env = os.Environ()
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		checkAndExit(cmd.Run())
	} else {
		for _, bin := range binPaths {
			fmt.Printf("👉 remove %s\n", bin)
			_ = os.Remove(bin)
		}
		fmt.Printf("👉 remove %s\n", filepath.Join(dir, "mmh"))
		_ = os.Remove(filepath.Join(dir, "mmh"))
	}
}
