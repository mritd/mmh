package core

import (
	"fmt"
	"io"
	"os"
	osexec "os/exec"
	"path/filepath"

	"github.com/mritd/mmh/pkg/common"
)

// Install Install mmh binary files into the specified dir and create alias soft links
func Install(dir string) {
	var binPaths []string
	for _, as := range Aliases {
		binPaths = append(binPaths, filepath.Join(dir, as))
	}

	currentPath, err := osexec.LookPath(os.Args[0])
	common.CheckAndExit(err)

	if !common.IsRoot() {
		cmds := append(os.Environ(), currentPath, "install", "--dir", dir)
		cmd := osexec.Command("sudo", cmds...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		common.CheckAndExit(cmd.Run())
	} else {
		Uninstall(dir)
		f, err := os.Open(currentPath)
		common.CheckAndExit(err)
		defer func() { _ = f.Close() }()

		target, err := os.OpenFile(filepath.Join(dir, "mmh"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
		common.CheckAndExit(err)
		defer func() { _ = target.Close() }()

		fmt.Printf("📥 install %s\n", filepath.Join(dir, "mmh"))
		_, err = io.Copy(target, f)
		common.CheckAndExit(err)
		for _, bin := range binPaths {
			fmt.Printf("📥 install %s\n", bin)
			common.CheckAndExit(os.Symlink(filepath.Join(dir, "mmh"), bin))
		}
	}

}

// Uninstall deletes mmh binary files and related soft links from the specified dir
func Uninstall(dir string) {
	var binPaths []string
	for _, as := range Aliases {
		binPaths = append(binPaths, filepath.Join(dir, as))
	}

	currentPath, err := osexec.LookPath(os.Args[0])
	common.CheckAndExit(err)

	if !common.IsRoot() {
		cmds := append(os.Environ(), currentPath, "uninstall", "--dir", dir)
		cmd := osexec.Command("sudo", cmds...)
		cmd.Env = os.Environ()
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		common.CheckAndExit(cmd.Run())
	} else {
		for _, bin := range binPaths {
			fmt.Printf("👉 remove %s\n", bin)
			_ = os.Remove(bin)
		}
		fmt.Printf("👉 remove %s\n", filepath.Join(dir, "mmh"))
		_ = os.Remove(filepath.Join(dir, "mmh"))
	}
}
