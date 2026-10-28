package core

import (
	"os"
	"strings"
	"testing"
)

func TestAliasBlock(t *testing.T) {
	block := aliasBlock("mmh")

	for _, name := range AliasNames {
		want := "alias " + name + "='mmh " + name + "'"
		if !strings.Contains(block, want) {
			t.Errorf("alias block missing %q:\n%s", want, block)
		}
	}
	if !strings.HasPrefix(block, aliasBlockBegin+"\n") {
		t.Errorf("alias block missing begin marker:\n%s", block)
	}
	if !strings.HasSuffix(block, aliasBlockEnd+"\n") {
		t.Errorf("alias block missing end marker:\n%s", block)
	}
}

func TestMergeAliasBlock(t *testing.T) {
	block := aliasBlock("mmh")

	tests := []struct {
		name    string
		content string
		want    string
		changed bool
	}{
		{
			name:    "empty file",
			content: "",
			want:    block,
			changed: true,
		},
		{
			name:    "append to file without trailing newline",
			content: "export FOO=bar",
			want:    "export FOO=bar\n\n" + block,
			changed: true,
		},
		{
			name:    "append to file with trailing newline",
			content: "export FOO=bar\n",
			want:    "export FOO=bar\n\n" + block,
			changed: true,
		},
		{
			name:    "already injected",
			content: "export FOO=bar\n\n" + block,
			want:    "export FOO=bar\n\n" + block,
			changed: false,
		},
		{
			name:    "replace stale block and keep trailing content",
			content: "export FOO=bar\n\n" + aliasBlock("old-mmh") + "export BAZ=qux\n",
			want:    "export FOO=bar\n\n" + block + "export BAZ=qux\n",
			changed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, changed, err := mergeAliasBlock(tt.content, block)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("merge result mismatch:\n got: %q\nwant: %q", got, tt.want)
			}
			if changed != tt.changed {
				t.Errorf("changed = %v, want %v", changed, tt.changed)
			}

			// merging the result again must be a no-op (idempotent)
			again, changedAgain, err := mergeAliasBlock(got, block)
			if err != nil {
				t.Fatalf("unexpected error on second merge: %v", err)
			}
			if changedAgain || again != got {
				t.Errorf("merge is not idempotent:\n first: %q\nsecond: %q", got, again)
			}
			if strings.Count(again, aliasBlockBegin) != 1 || strings.Count(again, aliasBlockEnd) != 1 {
				t.Errorf("expected exactly one alias block:\n%s", again)
			}
		})
	}
}

func TestMergeAliasBlockMalformed(t *testing.T) {
	block := aliasBlock("mmh")

	tests := []struct {
		name    string
		content string
	}{
		{name: "missing end marker", content: aliasBlockBegin + "\nalias mcp='mmh mcp'\n"},
		{name: "missing begin marker", content: "alias mcp='mmh mcp'\n" + aliasBlockEnd + "\n"},
		{name: "reversed markers", content: aliasBlockEnd + "\n" + aliasBlockBegin + "\n"},
		{name: "duplicate begin marker", content: aliasBlockBegin + "\n" + aliasBlockBegin + "\n" + aliasBlockEnd + "\n"},
		{name: "duplicate end marker", content: aliasBlockBegin + "\n" + aliasBlockEnd + "\n" + aliasBlockEnd + "\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, _, err := mergeAliasBlock(tt.content, block); err == nil {
				t.Error("expected an error for malformed markers, got nil")
			}
			if _, _, err := removeAliasBlock(tt.content); err == nil {
				t.Error("expected an error for malformed markers on remove, got nil")
			}
		})
	}
}

func TestRemoveAliasBlock(t *testing.T) {
	block := aliasBlock("mmh")

	tests := []struct {
		name    string
		content string
		want    string
		changed bool
	}{
		{
			name:    "block only",
			content: block,
			want:    "",
			changed: true,
		},
		{
			name:    "block appended to existing content",
			content: "export FOO=bar\n\n" + block,
			want:    "export FOO=bar\n",
			changed: true,
		},
		{
			name:    "block followed by user content",
			content: block + "export BAZ=qux\n",
			want:    "export BAZ=qux\n",
			changed: true,
		},
		{
			name:    "no block",
			content: "export FOO=bar\n",
			want:    "export FOO=bar\n",
			changed: false,
		},
		{
			name:    "empty file",
			content: "",
			want:    "",
			changed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, changed, err := removeAliasBlock(tt.content)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("remove result mismatch:\n got: %q\nwant: %q", got, tt.want)
			}
			if changed != tt.changed {
				t.Errorf("changed = %v, want %v", changed, tt.changed)
			}

			// removing again must be a no-op
			again, changedAgain, err := removeAliasBlock(got)
			if err != nil {
				t.Fatalf("unexpected error on second remove: %v", err)
			}
			if changedAgain || again != got {
				t.Errorf("remove is not idempotent:\n first: %q\nsecond: %q", got, again)
			}
		})
	}
}

func TestDetectShell(t *testing.T) {
	old, existed := os.LookupEnv("SHELL")
	defer func() {
		if existed {
			_ = os.Setenv("SHELL", old)
		} else {
			_ = os.Unsetenv("SHELL")
		}
	}()

	tests := map[string]string{
		"/bin/zsh":            "zsh",
		"/usr/local/bin/bash": "bash",
		"/usr/bin/fish":       "fish",
		"/bin/sh":             "sh",
		"/bin/ksh":            "sh",
		"":                    "sh",
	}
	for shell, want := range tests {
		if err := os.Setenv("SHELL", shell); err != nil {
			t.Fatal(err)
		}
		if got := DetectShell(); got != want {
			t.Errorf("DetectShell() with SHELL=%q = %q, want %q", shell, got, want)
		}
	}
}
