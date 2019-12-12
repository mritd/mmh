package mmh

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"text/template"

	"github.com/mritd/promptx"

	"github.com/mitchellh/go-homedir"
)

const (
	ConfigDirEnvName       = "MMH_CONFIG_DIR"
	CurrentConfigStoreFile = ".current"
)

var (
	configOnce        sync.Once
	ConfigDir         string
	ConfigList        ConfigInfo
	BasicConfig       Config
	BasicConfigName   string
	BasicConfigPath   string
	CurrentConfig     Config
	CurrentConfigName string
	CurrentConfigPath string
	Aliases           []string
)

func LoadConfig() {
	configOnce.Do(func() {
		// get user home dir
		home, err := homedir.Dir()
		checkAndExit(err)
		// load config dir from env
		ConfigDir = os.Getenv(ConfigDirEnvName)
		if ConfigDir == "" {
			// default to $HOME/.mmh
			ConfigDir = filepath.Join(home, ".mmh")
			_, err = os.Stat(ConfigDir)
			if err != nil {
				if os.IsNotExist(err) {
					// create config dir
					checkAndExit(os.MkdirAll(ConfigDir, 0755))
					// create default config file
					CurrentConfigName = "default.yaml"
					CurrentConfigPath = filepath.Join(ConfigDir, CurrentConfigName)
					checkAndExit(ConfigExample().WriteTo(CurrentConfigPath))
					// create basic config file
					BasicConfigName = "basic.yaml"
					BasicConfigPath = filepath.Join(ConfigDir, BasicConfigName)
					checkAndExit(ConfigExample().WriteTo(BasicConfigPath))
					// set current config to default
					currentCfgStoreFile := filepath.Join(ConfigDir, CurrentConfigStoreFile)
					checkAndExit(ioutil.WriteFile(currentCfgStoreFile, []byte(CurrentConfigName), 0644))
				} else if err != nil {
					checkAndExit(err)
				}
			}
		}

		// check config dir if it not exist
		_, err = os.Stat(ConfigDir)
		checkAndExit(err)
		// config dir path only support absolute path or start with homedir(~)
		if !filepath.IsAbs(ConfigDir) && !strings.HasPrefix(ConfigDir, "~") {
			Exit("the config dir path must be a absolute path or start with homedir(~)", 1)
		}
		// convert config dir path with homedir(~) prefix to absolute path
		if strings.HasPrefix(ConfigDir, "~") {
			ConfigDir = strings.Replace(ConfigDir, "~", home, 1)
		}
		// get current config
		currentCfgStoreFile := filepath.Join(ConfigDir, CurrentConfigStoreFile)
		bs, err := ioutil.ReadFile(currentCfgStoreFile)
		if err != nil || len(bs) < 1 {
			fmt.Printf("failed to get current config, use default config\n")
			CurrentConfigName = "default.yaml"
		} else {
			CurrentConfigName = string(bs)
		}
		// load current config
		CurrentConfigPath = filepath.Join(ConfigDir, CurrentConfigName)
		checkErr(CurrentConfig.LoadFrom(CurrentConfigPath))
		// load basic config if it exist
		BasicConfigName = "basic.yaml"
		BasicConfigPath = filepath.Join(ConfigDir, BasicConfigName)
		if _, err = os.Stat(BasicConfigPath); err == nil {
			checkErr(BasicConfig.LoadFrom(BasicConfigPath))
		}

		// load all config info
		_ = filepath.Walk(ConfigDir, func(path string, f os.FileInfo, err error) error {
			if f.IsDir() || !strings.HasSuffix(f.Name(), ".yaml") {
				return nil
			}
			ConfigList = append(ConfigList, struct {
				Name      string
				Path      string
				IsCurrent bool
			}{
				Name:      strings.TrimSuffix(f.Name(), ".yaml"),
				Path:      path,
				IsCurrent: path == CurrentConfigPath,
			})
			return nil
		})
		sort.Sort(ConfigList)
	})
}

func ListConfig() {
	tpl := `  Name          Path
---------------------------------
{{ range . }}{{ if .IsCurrent }}{{ "» " | cyan }}{{ .Name | listLayout | cyan }}{{ else }}  {{ .Name | listLayout }}{{ end }}{{ if .IsCurrent }}{{ .Path | cyan }}{{ else }}{{ .Path }}{{ end }}
{{ end }}`

	funcMap := promptx.FuncMap
	funcMap["listLayout"] = listLayout

	t := template.New("").Funcs(funcMap)
	_, _ = t.Parse(tpl)

	var buf bytes.Buffer
	checkAndExit(t.Execute(&buf, ConfigList))
	fmt.Println(buf.String())
}

func SetConfig(name string) {
	// check config name exist
	hasConfig := false
	for _, c := range ConfigList {
		if c.Name == name {
			hasConfig = true
		}
	}
	if !hasConfig {
		Exit(fmt.Sprintf("config [%s] not exist", name), 1)
	}
	// write
	checkAndExit(ioutil.WriteFile(filepath.Join(ConfigDir, CurrentConfigStoreFile), []byte(name+".yaml"), 0644))
}

func InteractiveSetConfig() {
	cfg := &promptx.SelectConfig{
		ActiveTpl:    `»  {{ .Name | cyan }}`,
		InactiveTpl:  `  {{ .Name | white }}`,
		SelectPrompt: "Config",
		SelectedTpl:  `{{ "» " | green }}{{ .Name | green }}`,
		DisPlaySize:  9,
		DetailsTpl: `
--------- Context ----------
{{ "Name:" | faint }} {{ .Name | faint }}
{{ "Path:" | faint }} {{ .Path | faint }}`,
	}

	idx := (&promptx.Select{Items: ConfigList, Config: cfg}).Run()
	SetConfig(strings.TrimSuffix(ConfigList[idx].Name, ".yaml"))
}
