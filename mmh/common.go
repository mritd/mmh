package mmh

import (
	"os"
	"os/user"
	"strings"

	"github.com/spf13/cobra"

	"fmt"
)

// print layout func
func listLayout(name string) string {
	if len(name) < 14 {
		return fmt.Sprintf("%-14s", name)
	} else {
		return fmt.Sprintf("%-14s", ShortenString(name, 14))
	}
}

// merge tags
func mergeTags(tags []string) string {
	return strings.Join(tags, ",")
}

func checkAndExit(err error) {
	if err != nil {
		fmt.Println("😱 " + err.Error())
		os.Exit(1)
	}
}

func checkErr(err error) bool {
	if err != nil {
		fmt.Println("😱 " + err.Error())
		return false
	}

	return true
}

func ShortenString(str string, n int) string {
	if len(str) <= n {
		return str
	} else {
		return str[:n]
	}
}

func Exit(message string, code int) {
	if strings.TrimSpace(message) == "" {
		message = "No message"
	}
	fmt.Println("😱 " + message)
	os.Exit(code)
}

func isRoot() bool {
	u, err := user.Current()
	checkAndExit(err)
	return u.Uid == "0" || u.Gid == "0"
}

func FindAllAliases(cmd *cobra.Command) []string {
	var aliases []string
	if cmd.HasSubCommands() {
		cmds := cmd.Commands()
		for _, c := range cmds {
			if len(c.Aliases) > 0 {
				aliases = append(aliases, c.Aliases...)
			}
			if c.HasSubCommands() {
				as := FindAllAliases(c)
				if len(as) > 0 {
					aliases = append(aliases, as...)
				}
			}
		}
	} else {
		if len(cmd.Aliases) > 0 {
			aliases = append(aliases, cmd.Aliases...)
		}
	}

	return aliases
}
