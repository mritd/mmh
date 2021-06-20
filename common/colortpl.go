package common

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"sync"
	"text/template"

	"github.com/fatih/color"
)

const renderTpl = `{{ %s .Prefix }} {{ .Value }}`

const (
	ColorRed     = "red"
	ColorGreen   = "green"
	ColorYellow  = "yellow"
	ColorBlue    = "blue"
	ColorMagenta = "magenta"
	ColorCyan    = "cyan"
	ColorWhite   = "white"
)

var colorOnce sync.Once
var tplCh = make(chan *template.Template, 1)
var tplCacheMux sync.RWMutex
var tplCacheMap = make(map[string]*template.Template, 7)

var colorFuncMap = template.FuncMap{
	ColorRed:     color.New(color.FgRed).SprintfFunc(),
	ColorGreen:   color.New(color.FgGreen).SprintfFunc(),
	ColorYellow:  color.New(color.FgYellow).SprintfFunc(),
	ColorBlue:    color.New(color.FgBlue).SprintfFunc(),
	ColorMagenta: color.New(color.FgMagenta).SprintfFunc(),
	ColorCyan:    color.New(color.FgCyan).SprintfFunc(),
	ColorWhite:   color.New(color.FgWhite).SprintfFunc(),
}

func RenderedTpl() *template.Template {
	colorOnce.Do(func() {
		go func() {
			for {
				for k := range colorFuncMap {
					tpl, _ := template.New("").Funcs(colorFuncMap).Parse(fmt.Sprintf(renderTpl, k))
					tplCh <- tpl
				}
			}
		}()
	})
	return <-tplCh
}

type ColorLine struct {
	Prefix string
	Value  string
}

func RenderedOutput(wr io.Writer, line ColorLine) error {
	colorTpl, ok := tplCacheMap[line.Prefix]
	if !ok {
		colorTpl = RenderedTpl()
		tplCacheMux.Lock()
		tplCacheMap[line.Prefix] = colorTpl
		tplCacheMux.Unlock()
	}
	return colorTpl.Execute(wr, line)
}

func Converted2Rendered(r io.Reader, w io.Writer, prefix string) {
	reader := bufio.NewReader(r)
	// use buf to ensure atomic output of each line
	var buf bytes.Buffer
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		_ = RenderedOutput(&buf, ColorLine{Prefix: prefix, Value: line})
		_, _ = io.Copy(w, &buf)
		buf.Reset()
	}
}
