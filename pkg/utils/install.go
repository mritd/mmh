/*
 * Copyright 2018 mritd <mritd1234@gmail.com>.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
