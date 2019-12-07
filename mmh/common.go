package mmh

import (
	"strings"

	"fmt"

	"github.com/mritd/mmh/utils"
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
