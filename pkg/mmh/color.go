package mmh

import (
	"fmt"
	"strconv"
	"strings"

	"text/template"
)

const esc = "\033["

type attribute int

// Foreground weight/decoration attributes.
const (
	reset attribute = iota
)

// Foreground color attributes
const (
	FGRed attribute = iota + 31
	FGGreen
	FGYellow
	FGBlue
	FGMagenta
	FGCyan
)

// ResetCode is the character code used to reset the terminal formatting
var ResetCode = fmt.Sprintf("%s%dm", esc, reset)

// ColorsFuncMap defines template helpers for the output. It can be extended as a
// regular map.
var ColorsFuncMap = template.FuncMap{
	"red":     Styler(FGRed),
	"green":   Styler(FGGreen),
	"yellow":  Styler(FGYellow),
	"blue":    Styler(FGBlue),
	"magenta": Styler(FGMagenta),
	"cyan":    Styler(FGCyan),
}

// Styler returns a func that applies the attributes given in the Styler call
// to the provided string.
func Styler(attrs ...attribute) func(interface{}) string {
	attrstrs := make([]string, len(attrs))
	for i, v := range attrs {
		attrstrs[i] = strconv.Itoa(int(v))
	}

	seq := strings.Join(attrstrs, ";")

	return func(v interface{}) string {
		end := ""
		s, ok := v.(string)
		if !ok || !strings.HasSuffix(s, ResetCode) {
			end = ResetCode
		}
		return fmt.Sprintf("%s%sm%v%s", esc, seq, v, end)
	}
}
