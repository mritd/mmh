package core

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/mritd/mmh/common"
)

const (
	aliasBlockBegin = "# >>> mmh aliases >>>"
	aliasBlockEnd   = "# <<< mmh aliases <<<"
)

// AliasNames is the list of mmh subcommands which will be injected into shell config files
var AliasNames = []string{"mcp", "mcs", "mcx", "mec", "mgo", "mping", "mtun"}

// supportedShells is the shell list which can be detected or specified via --shell
var supportedShells = []string{"zsh", "bash", "fish", "sh"}

// InstallAlias injects the mmh subcommand aliases into shell config files.
//
// shell is a comma separated shell list (zsh, bash, fish, sh); when it is empty the
// current shell is detected from $SHELL. file points to an arbitrary config file and
// takes precedence over shell. The aliases are always written between managed marker
// lines, so running this function multiple times only replaces the existing block
// instead of appending a new one. When remove is true the managed block is deleted.
func InstallAlias(shell, file, bin string, remove bool) {
	paths, err := aliasTargets(shell, file)
	common.CheckAndExit(err)

	if bin == "" {
		bin = "mmh"
	}
	block := aliasBlock(bin)

	for _, path := range paths {
		changed, err := updateAliasFile(path, block, remove)
		common.CheckAndExit(err)
		printAliasResult(path, changed, remove)
	}
}

// aliasTargets resolves the config file list which should be touched
func aliasTargets(shell, file string) ([]string, error) {
	if file != "" {
		return []string{expandHome(file)}, nil
	}

	var shells []string
	for _, s := range strings.Split(shell, ",") {
		if s = strings.ToLower(strings.TrimSpace(s)); s != "" {
			shells = append(shells, s)
		}
	}
	if len(shells) == 0 {
		shells = []string{DetectShell()}
	}

	var paths []string
	seen := make(map[string]bool, len(shells))
	for _, s := range shells {
		path, err := shellRcFile(s)
		if err != nil {
			return nil, err
		}
		// different shell names may resolve to the same config file (e.g. bash/sh)
		if !seen[path] {
			seen[path] = true
			paths = append(paths, path)
		}
	}
	return paths, nil
}

// DetectShell returns the shell name of the current environment, unknown shells
// fall back to sh (the POSIX profile)
func DetectShell() string {
	shell := strings.ToLower(filepath.Base(os.Getenv("SHELL")))
	switch shell {
	case "zsh", "bash", "fish":
		return shell
	default:
		return "sh"
	}
}

// shellRcFile returns the config file which is read by the given shell
func shellRcFile(shell string) (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", err
	}

	switch shell {
	case "zsh":
		// ZDOTDIR overrides the default location of the zsh config files
		if dir := os.Getenv("ZDOTDIR"); dir != "" {
			return filepath.Join(expandHome(dir), ".zshrc"), nil
		}
		return filepath.Join(home, ".zshrc"), nil
	case "bash":
		return bashRcFile(home), nil
	case "fish":
		if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
			return filepath.Join(expandHome(dir), "fish", "config.fish"), nil
		}
		return filepath.Join(home, ".config", "fish", "config.fish"), nil
	case "sh":
		return filepath.Join(home, ".profile"), nil
	default:
		return "", fmt.Errorf("unsupported shell [%s], supported shells: %s", shell, strings.Join(supportedShells, ", "))
	}
}

// bashRcFile returns the bash config file which is actually read by the terminal.
// macOS terminals start login shells and read ~/.bash_profile, while most Linux
// terminals start interactive shells and read ~/.bashrc; the first existing file
// of the platform specific candidate list wins.
func bashRcFile(home string) string {
	candidates := []string{
		filepath.Join(home, ".bashrc"),
		filepath.Join(home, ".bash_profile"),
		filepath.Join(home, ".profile"),
	}
	if runtime.GOOS == "darwin" {
		candidates = []string{
			filepath.Join(home, ".bash_profile"),
			filepath.Join(home, ".bashrc"),
			filepath.Join(home, ".profile"),
		}
	}
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return candidates[0]
}

// aliasBlock builds the managed alias block, the alias syntax is valid in
// bash/zsh/sh and fish
func aliasBlock(bin string) string {
	var b strings.Builder
	b.WriteString(aliasBlockBegin + "\n")
	b.WriteString("# managed by mmh, changes will be overwritten\n")
	for _, name := range AliasNames {
		fmt.Fprintf(&b, "alias %s='%s %s'\n", name, bin, name)
	}
	b.WriteString(aliasBlockEnd + "\n")
	return b.String()
}

// updateAliasFile injects or removes the alias block, it reports whether the
// file content was changed
func updateAliasFile(path, block string, remove bool) (bool, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return false, err
		}
		// nothing to remove, but the config file should be created when installing
		if remove {
			return false, nil
		}
	}

	var result string
	var changed bool
	if remove {
		result, changed, err = removeAliasBlock(string(content))
	} else {
		result, changed, err = mergeAliasBlock(string(content), block)
	}
	if err != nil {
		return false, fmt.Errorf("%s: %w", path, err)
	}
	if !changed {
		return false, nil
	}

	mode := os.FileMode(0644)
	if info, statErr := os.Stat(path); statErr == nil {
		mode = info.Mode().Perm()
	} else if os.IsNotExist(statErr) {
		if err = os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return false, err
		}
	}
	return true, ioutil.WriteFile(path, []byte(result), mode)
}

// mergeAliasBlock replaces the managed block in content with block; when no
// managed block exists, block is appended to the end of content
func mergeAliasBlock(content, block string) (string, bool, error) {
	lines := strings.Split(content, "\n")
	begin, end, err := scanAliasBlock(lines)
	if err != nil {
		return "", false, err
	}

	var result string
	if begin == -1 {
		var b strings.Builder
		b.WriteString(content)
		if content != "" && !strings.HasSuffix(content, "\n") {
			b.WriteString("\n")
		}
		if content != "" {
			b.WriteString("\n")
		}
		b.WriteString(block)
		result = b.String()
	} else {
		result = blockPrefix(lines[:begin]) + block + strings.Join(lines[end+1:], "\n")
	}
	return result, result != content, nil
}

// removeAliasBlock deletes the managed block from content, content is returned
// unchanged when no managed block exists
func removeAliasBlock(content string) (string, bool, error) {
	lines := strings.Split(content, "\n")
	begin, end, err := scanAliasBlock(lines)
	if err != nil {
		return "", false, err
	}
	if begin == -1 {
		return content, false, nil
	}

	prefix := blockPrefix(lines[:begin])
	// the block is inserted with a blank separator line, drop it together with the block
	if strings.HasSuffix(prefix, "\n\n") {
		prefix = strings.TrimSuffix(prefix, "\n")
	}
	result := prefix + strings.Join(lines[end+1:], "\n")
	return result, result != content, nil
}

// blockPrefix renders the lines before the managed block back to text, the
// newline which separates them from the following line must be kept
func blockPrefix(lines []string) string {
	prefix := strings.Join(lines, "\n")
	if len(lines) > 0 {
		prefix += "\n"
	}
	return prefix
}

// scanAliasBlock finds the managed block line range, -1 means the marker is absent
func scanAliasBlock(lines []string) (int, int, error) {
	begin, end := -1, -1
	for i, line := range lines {
		switch strings.TrimSpace(line) {
		case aliasBlockBegin:
			if begin != -1 {
				return -1, -1, fmt.Errorf("duplicate %q marker", aliasBlockBegin)
			}
			begin = i
		case aliasBlockEnd:
			if end != -1 {
				return -1, -1, fmt.Errorf("duplicate %q marker", aliasBlockEnd)
			}
			end = i
		}
	}

	switch {
	case begin == -1 && end == -1:
		return -1, -1, nil
	case begin == -1 || end == -1 || end < begin:
		return -1, -1, fmt.Errorf("malformed mmh alias block: %q must come before %q", aliasBlockBegin, aliasBlockEnd)
	}
	return begin, end, nil
}

// expandHome converts a path with the ~ prefix to the absolute path
func expandHome(path string) string {
	if strings.HasPrefix(path, "~") {
		if home, err := homedir.Dir(); err == nil {
			return strings.Replace(path, "~", home, 1)
		}
	}
	return path
}

func printAliasResult(path string, changed, remove bool) {
	switch {
	case remove && changed:
		fmt.Printf("🗑  removed mmh aliases from %s\n", path)
	case remove:
		fmt.Printf("🤷 no mmh aliases found in %s\n", path)
	case changed:
		fmt.Printf("✅ mmh aliases injected into %s\n", path)
		fmt.Printf("   run `source %s` or open a new terminal to take effect\n", path)
	default:
		fmt.Printf("✅ mmh aliases already up to date in %s\n", path)
	}
}
