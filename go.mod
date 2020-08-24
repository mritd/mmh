module github.com/mritd/mmh

go 1.14

require (
	github.com/fatih/color v1.9.0
	github.com/gorilla/mux v1.7.4
	github.com/json-iterator/go v1.1.10
	github.com/mitchellh/go-homedir v1.1.0
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/mritd/promptx v0.0.0-20200515061936-66e281bd3c15
	github.com/mritd/touchid v0.0.0-20200824095859-29eb0605b0ed
	github.com/pkg/sftp v1.11.0
	github.com/spf13/cobra v1.0.0
	github.com/xyproto/clip v0.3.1
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	gopkg.in/yaml.v2 v2.3.0
)

replace github.com/xyproto/clip v0.3.1 => github.com/mritd/clip v0.3.2-0.20200817040708-ed826a857db0
