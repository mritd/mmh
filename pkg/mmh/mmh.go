/*
 * Copyright 2018 mritd <mritd1234@gmail.com>.
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

import (
	"fmt"

	"sort"

	"github.com/mritd/mmh/pkg/utils"
	"github.com/mritd/promptx"
	"github.com/spf13/viper"
)

func SingleLogin(name string) {
	s := findServerByName(name)
	if s == nil {
		fmt.Println("Server not found!")
	} else {
		s.Connect()
	}
}

func InteractiveLogin() {
	var servers Servers
	utils.CheckAndExit(viper.UnmarshalKey("servers", &servers))
	sort.Sort(servers)

	cfg := &promptx.SelectConfig{
		ActiveTpl:    `»  {{ .Name | cyan }}: {{ .User | cyan }}{{ "@" | cyan }}{{ .Address | cyan }}`,
		InactiveTpl:  `  {{ .Name | white }}: {{ .User | white }}{{ "@" | white }}{{ .Address | white }}`,
		SelectPrompt: "Login Server",
		SelectedTpl:  `{{ "» " | green }}{{ .Name | green }}: {{ .User | green }}{{ "@" | green }}{{ .Address | green }}`,
		DisPlaySize:  9,
		DetailsTpl: `
--------- Login Server ----------
{{ "Name:" | faint }} {{ .Name | faint }}
{{ "User:" | faint }} {{ .User | faint }}
{{ "Address:" | faint }} {{ .Address | faint }}{{ ":" | faint }}{{ .Port | faint }}`,
	}

	s := &promptx.Select{
		Items:  servers,
		Config: cfg,
	}
	idx := s.Run()
	SingleLogin(servers[idx].Name)
}
