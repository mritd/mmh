package mmh

import (
	"errors"
	"strings"

	"fmt"

	"github.com/mritd/mmh/utils"
)

// config
var (
	Main           MainConfig
	BasicContext   ContextConfig
	CurrentContext ContextConfig
	MaxProxy       int
)

// error def
var (
	inputEmptyErr    = errors.New("input is empty")
	inputTooLongErr  = errors.New("input length must be <= 12")
	serverExistErr   = errors.New("server name exist")
	notNumberErr     = errors.New("only number support")
	proxyNotFoundErr = errors.New("proxy server not found")
)

// print layout func
func listLayout(name string) string {
	if len(name) < 12 {
		return fmt.Sprintf("%-12s", name)
	} else {
		return fmt.Sprintf("%-12s", utils.ShortenString(name, 12))
	}
}

// merge tags
func mergeTags(tags []string) string {
	return strings.Join(tags, ",")
}
