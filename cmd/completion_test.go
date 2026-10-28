package cmd

import (
	"strings"
	"testing"

	"github.com/mitchellh/go-homedir"
	"github.com/mritd/mmh/core"
)

func TestGenCompletion(t *testing.T) {
	for _, shell := range completionShells {
		t.Run(shell, func(t *testing.T) {
			script, err := genCompletion(rootCmd, shell)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.Contains(script, "mmh") {
				t.Errorf("%s completion script does not mention mmh", shell)
			}
		})
	}

	if _, err := genCompletion(rootCmd, "tcsh"); err == nil {
		t.Error("expected an error for an unsupported shell, got nil")
	}
}

func TestRegisterAliasCompletion(t *testing.T) {
	const script = "fake script"

	tests := []struct {
		shell string
		want  string
	}{
		{shell: "zsh", want: "compdef _mmh"},
		{shell: "bash", want: "complete -o default -o nospace -F __start_mmh"},
		{shell: "fish", want: "complete -c"},
		{shell: "powershell", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.shell, func(t *testing.T) {
			got := registerAliasCompletion(tt.shell, script)
			if !strings.HasPrefix(got, ensureTrailingNewline(script)) {
				t.Errorf("original script was not kept:\n%s", got)
			}
			if !strings.HasSuffix(got, "\n") {
				t.Error("registered script should end with a newline")
			}

			if tt.want == "" {
				if got != ensureTrailingNewline(script) {
					t.Errorf("powershell script should be unchanged, got:\n%s", got)
				}
				return
			}
			if !strings.Contains(got, tt.want) {
				t.Errorf("missing %q in:\n%s", tt.want, got)
			}
			for _, name := range core.AliasNames {
				if !strings.Contains(got, name) {
					t.Errorf("missing alias %q in:\n%s", name, got)
				}
			}
		})
	}
}

func TestCompletionInstallPath(t *testing.T) {
	oldDisableCache := homedir.DisableCache
	homedir.DisableCache = true
	defer func() { homedir.DisableCache = oldDisableCache }()

	t.Setenv("HOME", "/home/mmh")
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("ZSH", "")
	t.Setenv("ZSH_CUSTOM", "")

	tests := []struct {
		name  string
		env   map[string]string
		shell string
		want  string
	}{
		{
			name:  "zsh default",
			shell: "zsh",
			want:  "/home/mmh/.zsh/completions/_mmh",
		},
		{
			name:  "zsh oh-my-zsh",
			env:   map[string]string{"ZSH": "/home/mmh/.oh-my-zsh"},
			shell: "zsh",
			want:  "/home/mmh/.oh-my-zsh/custom/completions/_mmh",
		},
		{
			name:  "zsh custom dir",
			env:   map[string]string{"ZSH_CUSTOM": "/home/mmh/.zsh-custom"},
			shell: "zsh",
			want:  "/home/mmh/.zsh-custom/completions/_mmh",
		},
		{
			name:  "bash default",
			shell: "bash",
			want:  "/home/mmh/.local/share/bash-completion/completions/mmh",
		},
		{
			name:  "bash xdg data home",
			env:   map[string]string{"XDG_DATA_HOME": "/data"},
			shell: "bash",
			want:  "/data/bash-completion/completions/mmh",
		},
		{
			name:  "fish default",
			shell: "fish",
			want:  "/home/mmh/.config/fish/completions/mmh.fish",
		},
		{
			name:  "fish xdg config home",
			env:   map[string]string{"XDG_CONFIG_HOME": "/conf"},
			shell: "fish",
			want:  "/conf/fish/completions/mmh.fish",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.env {
				t.Setenv(k, v)
			}
			got, err := completionInstallPath(tt.shell)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("completionInstallPath(%q) = %q, want %q", tt.shell, got, tt.want)
			}
		})
	}

	if _, err := completionInstallPath("powershell"); err == nil {
		t.Error("expected an error for powershell install, got nil")
	}
}

func TestEnsureTrailingNewline(t *testing.T) {
	if got := ensureTrailingNewline("a"); got != "a\n" {
		t.Errorf("got %q, want %q", got, "a\n")
	}
	if got := ensureTrailingNewline("a\n"); got != "a\n" {
		t.Errorf("got %q, want %q", got, "a\n")
	}
	if got := ensureTrailingNewline(""); got != "" {
		t.Errorf("got %q, want empty", got)
	}
}
