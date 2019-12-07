package mmh

import (
	"strings"

	"fmt"

	"github.com/mritd/mmh/utils"
)

// print layout func
func listLayout(name string) string {
	if len(name) < 14 {
		return fmt.Sprintf("%-14s", name)
	} else {
		return fmt.Sprintf("%-14s", utils.ShortenString(name, 14))
	}
}

// merge tags
func mergeTags(tags []string) string {
	return strings.Join(tags, ",")
}
