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

	"github.com/PagerDuty/go-pdagent/cmd/cmdutil"
	"github.com/PagerDuty/go-pdagent/cmd/integrations/nagios"
	"github.com/PagerDuty/go-pdagent/pkg/common"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var rootCmd *cobra.Command

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	config := cmdutil.NewConfig()

	rootCmd = NewRootCmd(config)
	cobra.OnInitialize(initConfig)
}

func NewRootCmd(config *cmdutil.Config) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "pdagent",
		Short: "PagerDuty Agent CLI",
		Long: `A PagerDuty Agent and corresponding Command Line Interface.

	The agent acts as a local server between your own infrastructure and PagerDuty,
	providing command line tools to send PagerDuty events while ensuring event
	ordering and mitigating backpressure.

	On first run it's recommended you run "init" to generate a default
	configuration, then run "server" to start the agent itself.`,
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	defaults := cmdutil.GetDefaults()
	rootCmd.Version = common.Version

	pflags := rootCmd.PersistentFlags()
	pflags.StringVar(&cfgFile, "config", "", "config file (default is $HOME/.go-pdagent.yaml)")
	pflags.StringP("address", "a", defaults.Address, "address to run and access the agent server on.")
	pflags.String("pidfile", defaults.Pidfile, "pidfile for the currently running pdagent instance, if any.")
	pflags.StringP("secret", "s", defaults.Secret, "secret used to authorize agent access.")

	if err := viper.BindPFlag("address", pflags.Lookup("address")); err != nil {
		fmt.Println(err)
	}

	if err := viper.BindPFlag("pidfile", pflags.Lookup("pidfile")); err != nil {
		fmt.Println(err)
	}

	if err := viper.BindPFlag("secret", pflags.Lookup("secret")); err != nil {
		fmt.Println(err)
	}

	// All top-level commands go here
	rootCmd.AddCommand(NewEnqueueCmd(config))
	rootCmd.AddCommand(NewInitCmd())
	rootCmd.AddCommand(NewQueueCmd(config))
	rootCmd.AddCommand(NewSendCmd(config))
	rootCmd.AddCommand(NewServerCmd())
	rootCmd.AddCommand(NewVersionCmd())
	rootCmd.AddCommand(nagios.NewNagiosCmd(config))

	return rootCmd
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		// We add both production and dev paths here such that either config
		// will be automatically picked up.
		viper.AddConfigPath("/etc/pdagent/")
		viper.AddConfigPath(cmdutil.GetDefaultConfigPath())
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	_ = viper.ReadInConfig()
}
