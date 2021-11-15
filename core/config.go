package core

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/olekukonko/tablewriter"

	"github.com/mitchellh/go-homedir"
	"github.com/mritd/mmh/common"
)

const (
	// EnvConfigDirName The MMH_CONFIG_DIR env specifies the dir where the mmh config file is stored
	EnvConfigDirName = "MMH_CONFIG_DIR"

	ConfigNameFile  = ".current"
	BasicConfigName = "basic.yaml"
)

var (
	Configs ConfigList

	configDir string

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
			if err != nil {
				if !os.IsNotExist(err) {
					common.Exit(err.Error(), 1)
				}
				initConfig(configDir)
			}
		}
	}

	// get current config
	bs, err := ioutil.ReadFile(filepath.Join(configDir, ConfigNameFile))
	if err != nil || len(bs) < 1 {
		fmt.Println("failed to get current config, use default config(default.yaml)")
		currentConfigName = "default.yaml"
	} else {
		currentConfigName = string(bs)
	}
	// load current config
	currentConfigPath = filepath.Join(configDir, currentConfigName)
	common.PrintErr(currentConfig.LoadFrom(currentConfigPath))
	// load basic config if it exist
	basicConfigPath = filepath.Join(configDir, BasicConfigName)
	if _, err = os.Stat(basicConfigPath); err == nil {
		common.PrintErr(basicConfig.LoadFrom(basicConfigPath))
	}

	// load all config info
	_ = filepath.Walk(configDir, func(path string, f os.FileInfo, err error) error {
		if !common.CheckErr(err) {
			return nil
		}
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".yaml") {
			return nil
		}
		Configs = append(Configs, ConfigInfo{
			Name:      strings.TrimSuffix(f.Name(), ".yaml"),
			Path:      path,
			IsCurrent: path == currentConfigPath,
		})
		return nil
	})
	sort.Sort(Configs)
}

// SetConfig set which config file to use, and writes the config file name into
// the file storage; the config file must exist or the operation fails
func SetConfig(name string) {
	// check config name exist
	var exist bool
	for _, c := range Configs {
		if c.Name == name {
			exist = true
		}
	}
	if !exist {
		common.Exit(fmt.Sprintf("config [%s] not exist", name), 1)
	}
	// write to file
	common.CheckAndExit(ioutil.WriteFile(filepath.Join(configDir, ConfigNameFile), []byte(name+".yaml"), 0644))
}

func ListConfigs() {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Path"})
	for _, c := range Configs {
		if c.Name+".yaml" == currentConfigName {
			table.Append([]string{fmt.Sprintf("\033[1;32m%s\033[0m", c.Name), c.Path})
		} else {
			table.Append([]string{c.Name, c.Path})
		}
	}
	table.Render()
}

// initConfig init the example config
func initConfig(dir string) {
	// create config dir
	common.CheckAndExit(os.MkdirAll(dir, 0755))
	// create basic config file
	common.CheckAndExit(ConfigExample().WriteTo(filepath.Join(dir, BasicConfigName)))
	// set current config to default
	common.CheckAndExit(ioutil.WriteFile(filepath.Join(dir, ConfigNameFile), []byte(BasicConfigName), 0644))
}
