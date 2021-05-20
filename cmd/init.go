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
package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/PagerDuty/go-pdagent/pkg/common"
	"github.com/spf13/viper"

	"github.com/spf13/cobra"
)

func NewInitCmd() *cobra.Command {
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Generate a new initial configuration file.",
		Long: `Generate a new initial configuration file

	Can be run without options to automatically generate defaults, or will use
	configuration options or an existing config as its basis.`,
		Run: func(cmd *cobra.Command, args []string) {
			defaults := getDefaults()
			configFile := path.Join(defaults.ConfigPath, "config.yaml")

			if common.IsProduction() {
				fmt.Printf("Generating production config to %v\n", configFile)
			} else {
				fmt.Printf("Generating config to %v\n", configFile)
			}

			if err := os.MkdirAll(path.Dir(configFile), 0744); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			viper.SetConfigType("yaml")

			if err := viper.SafeWriteConfigAs(configFile); err != nil {
				fmt.Printf("Error writing config: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Config file generated to %v\n", configFile)
		},
	}

	return initCmd
}
