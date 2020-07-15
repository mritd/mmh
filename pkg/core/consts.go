package core

const (
	listConfigTpl = `  Name          Path
---------------------------------
{{ range . }}{{ if .IsCurrent }}{{ "» " | cyan }}{{ maxLen 15 .Name | cyan }}{{ else }}  {{ maxLen 15 .Name }}{{ end }}{{ if .IsCurrent }}{{ .Path | cyan }}{{ else }}{{ .Path }}{{ end }}
{{ end }}`

	listServersTpl = `Name           User      Tags                Address
-----------------------------------------------------------------
{{range . }}{{ maxLen 15 .Name }}{{ maxLen 10 .User }}{{ maxLen 20 (.Tags | mergeTags) }}{{ .Address }}:{{ .Port }}
{{end}}`

	serverDetailTpl = `Name: {{ .Name }}
User: {{ .User }}
Address: {{ .Address }}:{{ .Port }}
Tags: {{ .Tags | mergeTags }}
Proxy: {{ .Proxy }}`

	interactiveLoginSelectedTpl = `{{ "» " | green }}{{ .User | green }}{{ "@" | green }}{{ .Address | green }}`
	interactiveLoginActiveTpl   = `»  {{ maxLen 10 .Name | cyan }}>> {{ .User | cyan }}{{ "@" | cyan }}{{ .Address | cyan }}`
	interactiveLoginInactiveTpl = `  {{ maxLen 10 .Name | white }}>>  {{ .User | white }}{{ "@" | white }}{{ .Address | white }}`
	interactiveLoginDetailsTpl  = `
--------- Login Server ----------
{{ "Name:" | faint }} {{ .Name | faint }}
{{ "User:" | faint }} {{ .User | faint }}
{{ "Address:" | faint }} {{ .Address | faint }}{{ ":" | faint }}{{ .Port | faint }}`
)
