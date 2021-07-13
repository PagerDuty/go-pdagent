/*
Copyright © 2020 PagerDuty, Inc. <info@pagerduty.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmdutil

import (
	"fmt"
	"os"
	"path"

	"github.com/PagerDuty/go-pdagent/pkg/common"
	"github.com/mitchellh/go-homedir"
)

type Defaults struct {
	Address    string
	ConfigPath string
	Database   string
	Pidfile    string
	Secret     string
	Region     string
}

func GetDefaults() Defaults {
	prod := common.IsProduction()

	if prod {
		return Defaults{
			Address:    "127.0.0.1:49463",
			ConfigPath: "/etc/pdagent/",
			Database:   "/var/db/pdagent/pdagent.db",
			Pidfile:    "/var/run/pdagent/pidfile",
			Secret:     common.GenerateKey(),
			Region:     "us",
		}
	}

	configPath := getDefaultConfigPath()

	return Defaults{
		Address:    "127.0.0.1:49463",
		ConfigPath: configPath,
		Database:   path.Join(configPath, "pdagent.db"),
		Pidfile:    path.Join(configPath, "pidfile"),
		Secret:     common.GenerateKey(),
		Region:     "us",
	}
}

func getDefaultConfigPath() string {
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return path.Join(home, ".pdagent")
}
