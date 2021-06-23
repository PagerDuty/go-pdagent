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
package test

import (
	"bytes"
	"io"
	"os"

	"github.com/PagerDuty/go-pdagent/pkg/cmdutil"
	"github.com/spf13/viper"
)

func CaptureStdout(f func() error) (string, error) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)

	return buf.String(), err
}

// InitConfigForIntegrationsTesting initializes config for testing pdagent integrations
// that would normally get set in the `cmd` package init() function.
func InitConfigForIntegrationsTesting() {
	// Set defaults that would normally get set using the root command's persistent fllags
	viper.SetDefault("address", cmdutil.GetDefaults().Address)
	viper.SetDefault("pidfile", cmdutil.GetDefaults().Pidfile)
	viper.SetDefault("secret", cmdutil.GetDefaults().Secret)

	// Load normal config from environment variables and config files
	cmdutil.InitConfig()
}
