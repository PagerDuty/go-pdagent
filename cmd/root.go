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

	"github.com/PagerDuty/pagerduty-agent/pkg/common"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "pdagent",
	Short: "PagerDuty Agent CLI",
	Long: `A PagerDuty Agent and corresponding Command Line Interface.

The agent acts as a local server between your own infrastructure and PagerDuty,
providing command line tools to send PagerDuty events while ensuring event
ordering and mitigating backpressure.

On first run it's recommended you run "init" to generate a default
configuration, then run "server" to start the agent itself.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	pflags := rootCmd.PersistentFlags()

	pflags.StringVar(&cfgFile, "config", "", "config file (default is $HOME/.pagerduty-agent.yaml)")
	pflags.StringP("address", "a", "127.0.0.1:49463", "address to run and access the agent server on.")
	pflags.StringP("secret", "s", "undefined", "secret used to authorize agent access.")

	viper.BindPFlag("address", pflags.Lookup("address"))
	viper.BindPFlag("secret", pflags.Lookup("secret"))
	viper.SetDefault("address", "localhost:49463")
	viper.SetDefault("secret", common.GenerateKey())
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".pagerduty-agent" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".pagerduty-agent")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	_ = viper.ReadInConfig()
}
