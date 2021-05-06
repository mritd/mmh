module github.com/mritd/mmh

go 1.15

require (
	github.com/fatih/color v1.10.0
	github.com/gorilla/mux v1.8.0
	github.com/json-iterator/go v1.1.11
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mritd/promptx v0.0.0-20200515061936-66e281bd3c15
	github.com/mritd/touchid v0.0.0-20200914041600-145dfa05fb2b
	github.com/pkg/sftp v1.13.0
	github.com/spf13/cobra v1.1.2
	github.com/xyproto/clip v0.3.1
	golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad
	gopkg.in/yaml.v2 v2.4.0
)

replace github.com/xyproto/clip v0.3.1 => github.com/mritd/clip v0.3.2-0.20200817040708-ed826a857db0
