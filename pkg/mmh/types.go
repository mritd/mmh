/*
 * Copyright 2018 mritd <mritd1234@gmail.com>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package mmh

type Context struct {
	IsRemote      bool   `yaml:"is_remote" mapstructure:"is_remote"`
	RemoteAddress string `yaml:"remote_address" mapstructure:"remote_address"`
	ConfigPath    string `yaml:"config_path" mapstructure:"config_path"`
}

type Contexts map[string]Context

type contextDetail struct {
	Name           string
	ConfigPath     string
	CurrentContext bool
}

type contextDetails []contextDetail

func (cd contextDetails) Len() int {
	return len(cd)
}
func (cd contextDetails) Less(i, j int) bool {
	return cd[i].Name < cd[j].Name
}
func (cd contextDetails) Swap(i, j int) {
	cd[i], cd[j] = cd[j], cd[i]
}

type Basic struct {
	User               string `yaml:"user" mapstructure:"user"`
	Password           string `yaml:"password" mapstructure:"password"`
	PrivateKey         string `yaml:"privatekey" mapstructure:"privatekey"`
	PrivateKeyPassword string `yaml:"privatekey_password" mapstructure:"privatekey_password"`
	Port               int    `yaml:"port" mapstructure:"port"`
	Proxy              string `yaml:"proxy" mapstructure:"proxy"`
}

type Server struct {
	Name               string   `yaml:"name" mapstructure:"name"`
	Tags               []string `yaml:"tags" mapstructure:"tags"`
	User               string   `yaml:"user" mapstructure:"user"`
	Password           string   `yaml:"password" mapstructure:"password"`
	PrivateKey         string   `yaml:"privatekey" mapstructure:"privatekey"`
	PrivateKeyPassword string   `yaml:"privatekey_password" mapstructure:"privatekey_password"`
	Address            string   `yaml:"address" mapstructure:"address"`
	Port               int      `yaml:"port" mapstructure:"port"`
	Proxy              string   `yaml:"proxy" mapstructure:"proxy"`
	proxyCount         int
}

type Servers []*Server

func (servers Servers) Len() int {
	return len(servers)
}
func (servers Servers) Less(i, j int) bool {
	return servers[i].Name < servers[j].Name
}
func (servers Servers) Swap(i, j int) {
	servers[i], servers[j] = servers[j], servers[i]
}
