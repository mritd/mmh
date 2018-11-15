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

package mmh

import (
	"sort"
	"sync"
	"text/template"

	"github.com/fatih/color"
)

const (
	ColorRed     = "red"
	ColorGreen   = "green"
	ColorYellow  = "yellow"
	ColorBlue    = "blue"
	ColorMagenta = "magenta"
	ColorCyan    = "cyan"
	ColorWhite   = "white"
)

type colorCount struct {
	name  string
	count int
}

type colorCounts []colorCount

func (cs colorCounts) Len() int {
	return len(cs)
}
func (cs colorCounts) Less(i, j int) bool {
	return cs[i].count < cs[j].count
}
func (cs colorCounts) Swap(i, j int) {
	cs[i], cs[j] = cs[j], cs[i]
}

var colorMux sync.Mutex

var cs = colorCounts{
	colorCount{ColorRed, 0},
	colorCount{ColorGreen, 0},
	colorCount{ColorYellow, 0},
	colorCount{ColorBlue, 0},
	colorCount{ColorMagenta, 0},
	colorCount{ColorCyan, 0},
	colorCount{ColorWhite, 0},
}

var ColorsFuncMap = template.FuncMap{
	ColorRed:     color.New(color.FgRed).SprintfFunc(),
	ColorGreen:   color.New(color.FgGreen).SprintfFunc(),
	ColorYellow:  color.New(color.FgYellow).SprintfFunc(),
	ColorBlue:    color.New(color.FgBlue).SprintfFunc(),
	ColorMagenta: color.New(color.FgMagenta).SprintfFunc(),
	ColorCyan:    color.New(color.FgCyan).SprintfFunc(),
	ColorWhite:   color.New(color.FgWhite).SprintfFunc(),
}

func getColorFuncName() string {
	colorMux.Lock()
	defer colorMux.Unlock()
	sort.Sort(cs)
	cs[0].count++
	return cs[0].name
}
