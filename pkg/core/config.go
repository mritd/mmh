package core

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mritd/promptx"

	"github.com/mritd/mmh/pkg/common"

	"github.com/mitchellh/go-homedir"
)

const (
	configDirEnvName       = "MMH_CONFIG_DIR"
	configChgSelectEnvName = "MMH_CONFIG_CHANGED_SELECT"
	currentConfigStoreFile = ".current"
	basicConfigName        = "basic.yaml"
)

var (
	Aliases []string

	configDir  string
	configList ConfigInfo

	basicConfig       Config
	currentConfig     Config
	currentConfigName string
	currentConfigPath string
	basicConfigPath   string
)

func LoadConfig() {
	// get user home dir
	home, err := homedir.Dir()
	common.CheckAndExit(err)
	// load config dir from env
	configDir = os.Getenv(configDirEnvName)
	if configDir == "" {
		// default to $HOME/.mmh
		configDir = filepath.Join(home, ".mmh")
		_, err = os.Stat(configDir)
		if err != nil {
			if os.IsNotExist(err) {
				// create config dir
				common.CheckAndExit(os.MkdirAll(configDir, 0755))
				// create default config file
				currentConfigName = "default.yaml"
				currentConfigPath = filepath.Join(configDir, currentConfigName)
				common.CheckAndExit(ConfigExample().WriteTo(currentConfigPath))
				// create basic config file
				basicConfigPath = filepath.Join(configDir, basicConfigName)
				common.CheckAndExit(ConfigExample().WriteTo(basicConfigPath))
				// set current config to default
				currentCfgStoreFile := filepath.Join(configDir, currentConfigStoreFile)
				common.CheckAndExit(ioutil.WriteFile(currentCfgStoreFile, []byte(currentConfigName), 0644))
			} else if err != nil {
				common.Exit(err.Error(), 1)
			}
		}
	}

	// config dir path only support absolute path or start with homedir(~)
	if !filepath.IsAbs(configDir) && !strings.HasPrefix(configDir, "~") {
		common.Exit("the config dir path must be a absolute path or start with homedir(~)", 1)
	}
	// convert config dir path with homedir(~) prefix to absolute path
	if strings.HasPrefix(configDir, "~") {
		configDir = strings.Replace(configDir, "~", home, 1)
	}

	// check config dir if it not exist
	f, err := os.Lstat(configDir)
	common.CheckAndExit(err)

	// check config dir is symlink. filepath Walk does not follow symbolic links
	if f.Mode()&os.ModeSymlink != 0 {
		configDir, err = os.Readlink(configDir)
		common.CheckAndExit(err)
	}

	// get current config
	currentCfgStoreFile := filepath.Join(configDir, currentConfigStoreFile)
	bs, err := ioutil.ReadFile(currentCfgStoreFile)
	if err != nil || len(bs) < 1 {
		fmt.Println("failed to get current config, use default config")
		currentConfigName = "default.yaml"
	} else {
		currentConfigName = string(bs)
	}
	// load current config
	currentConfigPath = filepath.Join(configDir, currentConfigName)
	common.CheckErr(currentConfig.LoadFrom(currentConfigPath))
	// load basic config if it exist
	basicConfigPath = filepath.Join(configDir, basicConfigName)
	if _, err = os.Stat(basicConfigPath); err == nil {
		common.CheckErr(basicConfig.LoadFrom(basicConfigPath))
	}

	// load all config info
	_ = filepath.Walk(configDir, func(path string, f os.FileInfo, err error) error {
		if !common.CheckErr(err) {
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
}

func ReloadConfig() {
	configDir = ""
	configList = ConfigInfo{}

	basicConfig = Config{}
	currentConfig = Config{}
	currentConfigName = ""
	currentConfigPath = ""
	basicConfigPath = ""
	LoadConfig()
}

func ListConfig() {
	t, _ := common.Template(listConfigTpl)
	var buf bytes.Buffer
	common.CheckAndExit(t.Execute(&buf, configList))
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
		common.Exit(fmt.Sprintf("config [%s] not exist", name), 1)
	}
	// write
	common.CheckAndExit(ioutil.WriteFile(filepath.Join(configDir, currentConfigStoreFile), []byte(name+".yaml"), 0644))
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

	if os.Getenv(configChgSelectEnvName) == "true" {
		ReloadConfig()
		SingleInteractiveLogin()
	}
}
