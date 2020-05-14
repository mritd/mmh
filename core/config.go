package core

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
	configDirEnvName       = "MMH_CONFIG_DIR"
	currentConfigStoreFile = ".current"
	basicConfigName        = "basic.yaml"
)

var (
	Aliases []string

	configOnce sync.Once
	configDir  string
	configList ConfigInfo

	basicConfig       Config
	currentConfig     Config
	currentConfigName string
	currentConfigPath string
	basicConfigPath   string
)

func LoadConfig() {
	configOnce.Do(func() {
		// get user home dir
		home, err := homedir.Dir()
		checkAndExit(err)
		// load config dir from env
		configDir = os.Getenv(configDirEnvName)
		if configDir == "" {
			// default to $HOME/.mmh
			configDir = filepath.Join(home, ".mmh")
			_, err = os.Stat(configDir)
			if err != nil {
				if os.IsNotExist(err) {
					// create config dir
					checkAndExit(os.MkdirAll(configDir, 0755))
					// create default config file
					currentConfigName = "default.yaml"
					currentConfigPath = filepath.Join(configDir, currentConfigName)
					checkAndExit(ConfigExample().WriteTo(currentConfigPath))
					// create basic config file
					basicConfigPath = filepath.Join(configDir, basicConfigName)
					checkAndExit(ConfigExample().WriteTo(basicConfigPath))
					// set current config to default
					currentCfgStoreFile := filepath.Join(configDir, currentConfigStoreFile)
					checkAndExit(ioutil.WriteFile(currentCfgStoreFile, []byte(currentConfigName), 0644))
				} else if err != nil {
					Exit(err.Error(), 1)
				}
			}
		}

		// config dir path only support absolute path or start with homedir(~)
		if !filepath.IsAbs(configDir) && !strings.HasPrefix(configDir, "~") {
			Exit("the config dir path must be a absolute path or start with homedir(~)", 1)
		}
		// convert config dir path with homedir(~) prefix to absolute path
		if strings.HasPrefix(configDir, "~") {
			configDir = strings.Replace(configDir, "~", home, 1)
		}

		// check config dir if it not exist
		f, err := os.Lstat(configDir)
		if err != nil {
			return
		}

		// check config dir is symlink. filepath Walk does not follow symbolic links
		if f.Mode()&os.ModeSymlink != 0 {
			configDir, err = os.Readlink(configDir)
			checkAndExit(err)
		}

		// get current config
		currentCfgStoreFile := filepath.Join(configDir, currentConfigStoreFile)
		bs, err := ioutil.ReadFile(currentCfgStoreFile)
		if err != nil || len(bs) < 1 {
			fmt.Printf("failed to get current config, use default config\n")
			currentConfigName = "default.yaml"
		} else {
			currentConfigName = string(bs)
		}
		// load current config
		currentConfigPath = filepath.Join(configDir, currentConfigName)
		checkErr(currentConfig.LoadFrom(currentConfigPath))
		// load basic config if it exist
		basicConfigPath = filepath.Join(configDir, basicConfigName)
		if _, err = os.Stat(basicConfigPath); err == nil {
			checkErr(basicConfig.LoadFrom(basicConfigPath))
		}

		// load all config info
		_ = filepath.Walk(configDir, func(path string, f os.FileInfo, err error) error {
			if !checkErr(err) {
				return err
			}
			if f.IsDir() || !strings.HasSuffix(f.Name(), ".yaml") {
				return nil
			}
			configList = append(configList, struct {
				Name      string
				Path      string
				IsCurrent bool
			}{
				Name:      strings.TrimSuffix(f.Name(), ".yaml"),
				Path:      path,
				IsCurrent: path == currentConfigPath,
			})
			return nil
		})
		sort.Sort(configList)
	})
}

func ListConfig() {
	tpl := `  Name          Path
---------------------------------
{{ range . }}{{ if .IsCurrent }}{{ "» " | cyan }}{{ .Name | listLayout | cyan }}{{ else }}  {{ .Name | listLayout }}{{ end }}{{ if .IsCurrent }}{{ .Path | cyan }}{{ else }}{{ .Path }}{{ end }}
{{ end }}`

	funcMap := promptx.FuncMap
	funcMap["listLayout"] = listLayout
	t, _ := template.New("").Funcs(funcMap).Parse(tpl)

	var buf bytes.Buffer
	checkAndExit(t.Execute(&buf, configList))
	fmt.Println(buf.String())
}

func SetConfig(name string) {
	// check config name exist
	var exist bool
	for _, c := range configList {
		if c.Name == name {
			exist = true
		}
	}
	if !exist {
		Exit(fmt.Sprintf("config [%s] not exist", name), 1)
	}
	// write
	checkAndExit(ioutil.WriteFile(filepath.Join(configDir, currentConfigStoreFile), []byte(name+".yaml"), 0644))
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

	idx := (&promptx.Select{Items: configList, Config: cfg}).Run()
	SetConfig(strings.TrimSuffix(configList[idx].Name, ".yaml"))
}
