/*
 * Copyright 2018 mritd <mritd1234@gmail.com>
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
