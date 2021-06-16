module github.com/mritd/mmh

go 1.16

require (
	github.com/fatih/color v1.12.0
	github.com/gorilla/mux v1.8.0
	github.com/json-iterator/go v1.1.11
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mritd/touchid v0.0.1
	github.com/pkg/sftp v1.13.1
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.1.3
	github.com/xyproto/clip v0.3.1
	golang.org/x/crypto v0.0.0-20210421170649-83a5a9bb288b
	gopkg.in/yaml.v2 v2.4.0
)

replace github.com/xyproto/clip v0.3.1 => github.com/mritd/clip v0.3.2-0.20200817040708-ed826a857db0
