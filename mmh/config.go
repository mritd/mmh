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
	"github.com/mritd/mmh/utils"
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
)

func LoadConfig() {
	configOnce.Do(func() {
		// get user home dir
		home, err := homedir.Dir()
		utils.CheckAndExit(err)
		// load config dir from env
		ConfigDir = os.Getenv(ConfigDirEnvName)
		if ConfigDir == "" {
			// default to $HOME/.mmh
			ConfigDir = filepath.Join(home, ".mmh")
			_, err = os.Stat(ConfigDir)
			if err != nil {
				if os.IsNotExist(err) {
					// create config dir
					utils.CheckAndExit(os.MkdirAll(ConfigDir, 0755))
					// create default config file
					CurrentConfigName = "default.yaml"
					CurrentConfigPath = filepath.Join(ConfigDir, CurrentConfigName)
					utils.CheckAndExit(ConfigExample().WriteTo(CurrentConfigPath))
					// create basic config file
					BasicConfigName = "basic.yaml"
					BasicConfigPath = filepath.Join(ConfigDir, BasicConfigName)
					utils.CheckAndExit(ConfigExample().WriteTo(BasicConfigPath))
					// set current config to default
					currentCfgStoreFile := filepath.Join(ConfigDir, CurrentConfigStoreFile)
					utils.CheckAndExit(ioutil.WriteFile(currentCfgStoreFile, []byte(CurrentConfigName), 0644))
				} else if err != nil {
					utils.CheckAndExit(err)
				}
			}

		}

		// config dir path only support absolute path or start with homedir(~)
		if !filepath.IsAbs(ConfigDir) && !strings.HasPrefix(ConfigDir, "~") {
			utils.Exit("the config dir path must be a absolute path or start with homedir(~)", 1)
		}
		// convert config dir path with homedir(~) prefix to absolute path
		if strings.HasPrefix(ConfigDir, "~") {
			ConfigDir = strings.Replace(ConfigDir, "~", home, 1)
		}
		// get current config
		currentCfgStoreFile := filepath.Join(ConfigDir, CurrentConfigStoreFile)
		bs, err := ioutil.ReadFile(currentCfgStoreFile)
		if err != nil || len(bs) < 1 {
			fmt.Printf("failed to get current config, use [default.yaml]")
			CurrentConfigName = "default.yaml"
		} else {
			CurrentConfigName = string(bs)
		}
		// load current config
		CurrentConfigPath = filepath.Join(ConfigDir, CurrentConfigName)
		utils.CheckAndExit(CurrentConfig.LoadFrom(CurrentConfigPath))
		// load basic config if it exist
		BasicConfigName = "basic.yaml"
		BasicConfigPath = filepath.Join(ConfigDir, BasicConfigName)
		if _, err = os.Stat(BasicConfigPath); err == nil {
			utils.CheckAndExit(BasicConfig.LoadFrom(BasicConfigPath))
		}

		// load all config info
		_ = filepath.Walk(ConfigDir, func(path string, f os.FileInfo, err error) error {
			if f.IsDir() || f.Name() == ".current" {
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
{{ range . }}{{ if .IsCurrent }}{{ "» " | cyan }}{{ .Name | ListLayout | cyan }}{{ else }}  {{ .Name | ListLayout }}{{ end }}{{ if .IsCurrent }}{{ .Path | cyan }}{{ else }}{{ .Path }}{{ end }}
{{ end }}`

	funcMap := promptx.FuncMap
	funcMap["ListLayout"] = listLayout
	funcMap["MergeTag"] = mergeTags

	t := template.New("").Funcs(funcMap)
	_, _ = t.Parse(tpl)

	var buf bytes.Buffer
	utils.CheckAndExit(t.Execute(&buf, ConfigList))
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
		utils.Exit(fmt.Sprintf("config [%s] not exist", name), 1)
	}
	// write
	utils.CheckAndExit(ioutil.WriteFile(filepath.Join(ConfigDir, CurrentConfigStoreFile), []byte(name+".yaml"), 0644))
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
