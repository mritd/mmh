package core

const (
	listServersTpl = `Name           User      Tags                Address
-----------------------------------------------------------------
{{range . }}{{ maxLen 15 .Name }}{{ maxLen 10 .User }}{{ maxLen 20 (.Tags | mergeTags) }}{{ .Address }}:{{ .Port }}
{{end}}`

	serverDetailTpl = `Name: {{ .Name }}
User: {{ .User }}
Address: {{ .Address }}:{{ .Port }}
Tags: {{ .Tags | mergeTags }}
Proxy: {{ .Proxy }}
Config: {{ .ConfigPath }}`
)
