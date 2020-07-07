package common

import (
	"fmt"
	"io"
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

const defaultErrTpl = `{{ .Prefix | %s }}{{ ":" | %s }}  {{ .Value }}`

var colorOnce sync.Once
var colorCh = make(chan string, 1)
var colorMux sync.RWMutex
var colorOutMap = make(map[string]*template.Template, 5)

var colorFuncMap = template.FuncMap{
	ColorRed:     color.New(color.FgRed).SprintfFunc(),
	ColorGreen:   color.New(color.FgGreen).SprintfFunc(),
	ColorYellow:  color.New(color.FgYellow).SprintfFunc(),
	ColorBlue:    color.New(color.FgBlue).SprintfFunc(),
	ColorMagenta: color.New(color.FgMagenta).SprintfFunc(),
	ColorCyan:    color.New(color.FgCyan).SprintfFunc(),
	ColorWhite:   color.New(color.FgWhite).SprintfFunc(),
}

func colorName() string {
	colorOnce.Do(func() {
		go func() {
			for {
				for k := range colorFuncMap {
					colorCh <- k
				}
			}
		}()
	})
	return <-colorCh
}

type ColorLine struct {
	Prefix string
	Value  string
}

func ColorOutput(wr io.Writer, line ColorLine) error {
	colorTpl, ok := colorOutMap[line.Prefix]
	if !ok {
		colorName := colorName()
		tpl := fmt.Sprintf(defaultErrTpl, colorName, colorName)
		colorTpl, _ = template.New("").Funcs(colorFuncMap).Parse(tpl)
		colorMux.Lock()
		colorOutMap[line.Prefix] = colorTpl
		colorMux.Unlock()
	}
	return colorTpl.Execute(wr, line)
}
