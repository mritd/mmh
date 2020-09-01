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
	// The MMH_CONFIG_DIR env specifies the dir where the mmh config file is stored
	EnvConfigDirName = "MMH_CONFIG_DIR"
	// When the MMH_CONFIG_CHANGED_SELECT env is set to true, enter the interactive
	// server list after the config file is changed(mmh cf)
	EnvConfigChgSelectName = "MMH_CONFIG_CHANGED_SELECT"

	currentConfigStoreFile = ".current"
	basicConfigName        = "basic.yaml"
)

var (
	Aliases []string

	configDir  string
	configList ConfigList

	basicConfig       Config
	currentConfig     Config
	currentConfigName string
	currentConfigPath string
	basicConfigPath   string
)

// LoadConfig is responsible for loading config files and serializing them to memory objects
func LoadConfig() {
	// get user home dir
	home, err := homedir.Dir()
	common.CheckAndExit(err)

	// load config dir from env
	configDir = os.Getenv(EnvConfigDirName)
	if configDir != "" {
		// config dir path only support absolute path or start with homedir(~)
		if !filepath.IsAbs(configDir) && !strings.HasPrefix(configDir, "~") {
			common.Exit("the config dir path must be a absolute path or start with homedir(~)", 1)
		}
		// convert config dir path with homedir(~) prefix to absolute path
		if strings.HasPrefix(configDir, "~") {
			configDir = strings.Replace(configDir, "~", home, 1)
		}
	} else {
		// default to $HOME/.mmh
		configDir = filepath.Join(home, ".mmh")
	}

	// check config dir if it not exist
	f, err := os.Lstat(configDir)
	if err != nil {
		if os.IsNotExist(err) {
			initConfig(configDir)
		} else {
			common.Exit(err.Error(), 1)
		}
	} else {
		// check config dir is symlink. filepath Walk does not follow symbolic links
		if f.Mode()&os.ModeSymlink != 0 {
			configDir, err = os.Readlink(configDir)
			if os.IsNotExist(err) {
				initConfig(configDir)
			} else {
				common.Exit(err.Error(), 1)
			}
		}
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
	common.PrintErr(currentConfig.LoadFrom(currentConfigPath))
	// load basic config if it exist
	basicConfigPath = filepath.Join(configDir, basicConfigName)
	if _, err = os.Stat(basicConfigPath); err == nil {
		common.PrintErr(basicConfig.LoadFrom(basicConfigPath))
	}

	// load all config info
	_ = filepath.Walk(configDir, func(path string, f os.FileInfo, err error) error {
		if !common.CheckErr(err) {
			return err
		}
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".yaml") {
			return nil
		}
		configList = append(configList, ConfigInfo{
			Name:      strings.TrimSuffix(f.Name(), ".yaml"),
			Path:      path,
			IsCurrent: path == currentConfigPath,
		})
		return nil
	})
	sort.Sort(configList)
}

func initConfig(dir string) {
	// create config dir
	common.CheckAndExit(os.MkdirAll(dir, 0755))
	// create basic config file
	common.CheckAndExit(ConfigExample().WriteTo(filepath.Join(dir, basicConfigName)))
	// set current config to default
	common.CheckAndExit(ioutil.WriteFile(filepath.Join(dir, currentConfigStoreFile), []byte("basic.yaml"), 0644))
}

// ReloadConfig first clears the memory config objects, and then reloads them
func ReloadConfig() {
	configDir = ""
	configList = ConfigList{}

	basicConfig = Config{}
	currentConfig = Config{}
	currentConfigName = ""
	currentConfigPath = ""
	basicConfigPath = ""
	LoadConfig()
}

// ListConfig print a list of config files
func ListConfig() {
	t, _ := common.Template(listConfigTpl)
	var buf bytes.Buffer
	common.CheckAndExit(t.Execute(&buf, configList))
	fmt.Println(buf.String())
}

// SetConfig set which config file to use, and writes the config file name into
// the file storage; the config file must exist or the operation fails
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
	// write to file
	common.CheckAndExit(ioutil.WriteFile(filepath.Join(configDir, currentConfigStoreFile), []byte(name+".yaml"), 0644))
}

// InteractiveSetConfig provides interactive selection list based on SetConfig
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
	// If this variable is true, the interactive server list will be
	// displayed after the configuration file is changed
	if os.Getenv(EnvConfigChgSelectName) == "true" {
		ReloadConfig()
		SingleInteractiveLogin()
	}
}
