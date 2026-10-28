package core

import (
	"strings"
	"testing"
)

func renderLines(t *testing.T, groups []serverGroup) []string {
	t.Helper()

	out := renderServers(groups)
	if !strings.HasSuffix(out, "\n") {
		t.Fatalf("rendered output should end with a newline:\n%s", out)
	}
	return strings.Split(strings.TrimRight(out, "\n"), "\n")
}

func TestServerGroupName(t *testing.T) {
	tests := []struct {
		configPath string
		want       string
	}{
		{"/Users/walter/.mmh/basic.yaml", "basic"},
		{"/Users/walter/.mmh/hk-h20.yaml", "hk-h20"},
		{"", "-"},
	}
	for _, tt := range tests {
		g := serverGroup{configPath: tt.configPath}
		if got := g.name(); got != tt.want {
			t.Fatalf("serverGroup(%q).name() = %q, want %q", tt.configPath, got, tt.want)
		}
	}
}

func TestRenderServersShowsGroupColumn(t *testing.T) {
	groups := []serverGroup{
		{
			configPath: "/Users/walter/.mmh/basic.yaml",
			servers: Servers{
				{Name: "4090-1", User: "walter liu", Address: "192.168.212.21", Port: 22},
				{Name: "share", User: "root", Address: "share", Port: 22},
			},
		},
		{
			configPath: "/Users/walter/.mmh/hk-h20.yaml",
			servers: Servers{
				{Name: "gpu08", User: "walter", Address: "localhost", Port: 9833},
			},
		},
	}

	lines := renderLines(t, groups)
	// top border, header, header border, 2 rows, separator, 1 row, bottom border
	if len(lines) != 8 {
		t.Fatalf("expected 8 lines, got %d:\n%s", len(lines), strings.Join(lines, "\n"))
	}
	if !strings.Contains(lines[1], "GROUP") {
		t.Fatalf("header should contain a GROUP column:\n%s", strings.Join(lines, "\n"))
	}
	if lines[2] != lines[5] {
		t.Fatalf("separator line %q should equal the table border %q:\n%s", lines[5], lines[2], strings.Join(lines, "\n"))
	}
	if !strings.Contains(lines[3], "4090-1") || !strings.Contains(lines[3], "basic") {
		t.Fatalf("basic row should show its group:\n%s", strings.Join(lines, "\n"))
	}
	if !strings.Contains(lines[4], "share") || !strings.Contains(lines[4], "basic") {
		t.Fatalf("basic row should show its group:\n%s", strings.Join(lines, "\n"))
	}
	if !strings.Contains(lines[6], "gpu08") || !strings.Contains(lines[6], "hk-h20") {
		t.Fatalf("context row should show its group:\n%s", strings.Join(lines, "\n"))
	}
}

func TestRenderServersSingleGroupHasNoSeparator(t *testing.T) {
	groups := []serverGroup{
		{
			configPath: "/Users/walter/.mmh/basic.yaml",
			servers: Servers{
				{Name: "a", User: "root", Address: "10.0.0.1", Port: 22},
				{Name: "b", User: "root", Address: "10.0.0.2", Port: 22},
			},
		},
	}

	lines := renderLines(t, groups)
	// top border, header, header border, 2 rows, bottom border
	if len(lines) != 6 {
		t.Fatalf("expected 6 lines, got %d:\n%s", len(lines), strings.Join(lines, "\n"))
	}
}

func TestRenderServersSkipsEmptyGroups(t *testing.T) {
	groups := []serverGroup{
		{configPath: "/Users/walter/.mmh/basic.yaml"},
		{
			configPath: "/Users/walter/.mmh/hk-h20.yaml",
			servers: Servers{
				{Name: "a", User: "root", Address: "10.0.0.1", Port: 22},
			},
		},
	}

	lines := renderLines(t, groups)
	// top border, header, header border, 1 row, bottom border
	if len(lines) != 5 {
		t.Fatalf("expected 5 lines without a separator for the empty group, got %d:\n%s", len(lines), strings.Join(lines, "\n"))
	}
}

func TestRenderServersMultipleSeparators(t *testing.T) {
	groups := []serverGroup{
		{configPath: "/a.yaml", servers: Servers{{Name: "a", User: "root", Address: "10.0.0.1", Port: 22}}},
		{configPath: "/b.yaml", servers: Servers{{Name: "b", User: "root", Address: "10.0.0.2", Port: 22}}},
		{configPath: "/c.yaml", servers: Servers{{Name: "c", User: "root", Address: "10.0.0.3", Port: 22}}},
	}

	lines := renderLines(t, groups)
	borders := 0
	for _, l := range lines {
		if l == lines[2] {
			borders++
		}
	}
	// top border + header border + 2 separators + bottom border
	if borders != 5 {
		t.Fatalf("expected 5 border lines, got %d:\n%s", borders, strings.Join(lines, "\n"))
	}
}
