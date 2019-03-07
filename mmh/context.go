package mmh

import (
	"bytes"
	"fmt"
	"sort"
	"text/template"

	"github.com/mritd/promptx"

	"github.com/mritd/mmh/utils"
)

// find context by name
func FindContextByName(name string) (Context, bool) {
	for _, ctx := range Main.Contexts {
		if name == ctx.Name {
			return ctx, true
		}
	}
	return Context{}, false
}

// list contexts
func ListContexts() {

	tpl := `  Name          Path
---------------------------------
{{ range . }}{{ if .IsContext }}» {{ .Name | ListLayout }}{{ else }}  {{ .Name | ListLayout }}{{ end }}  {{ .ConfigPath }}
{{ end }}`

	t := template.New("").Funcs(map[string]interface{}{
		"ListLayout": listLayout,
		"MergeTag":   mergeTags,
	})
	_, _ = t.Parse(tpl)

	var ctxList []struct {
		Context
		IsContext bool
	}

	sort.Sort(Main.Contexts)
	for _, c := range Main.Contexts {
		ctxList = append(ctxList, struct {
			Context
			IsContext bool
		}{
			Context:   c,
			IsContext: c.Name == Main.Current})
	}

	var buf bytes.Buffer
	utils.CheckAndExit(t.Execute(&buf, ctxList))
	fmt.Println(buf.String())
}

// set current context
func SetContext(ctxName string) {
	_, ok := FindContextByName(ctxName)
	if !ok {
		utils.Exit(fmt.Sprintf("context [%s] not found", ctxName), 1)
	}
	Main.Current = ctxName
	utils.CheckAndExit(Main.Write())
}

func InteractiveSetContext() {
	cfg := &promptx.SelectConfig{
		ActiveTpl:    `»  {{ .Name | cyan }}`,
		InactiveTpl:  `  {{ .Name | white }}`,
		SelectPrompt: "Context",
		SelectedTpl:  `{{ "» " | green }}{{ .Name | green }}`,
		DisPlaySize:  9,
		DetailsTpl: `
--------- Context ----------
{{ "Name:" | faint }} {{ .Name | faint }}
{{ "ConfigPath:" | faint }} {{ .ConfigPath | faint }}`,
	}

	s := &promptx.Select{
		Items:  Main.Contexts,
		Config: cfg,
	}
	idx := s.Run()
	SetContext(Main.Contexts[idx].Name)
}
