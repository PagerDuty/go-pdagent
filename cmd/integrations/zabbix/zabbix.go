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
package zabbix

import (
	"github.com/PagerDuty/go-pdagent/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewZabbixCmd(config *cmdutil.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "zabbix",
		Short: "Access the Zabbix integration command(s).",
	}

	cmd.AddCommand(NewZabbixEnqueueCmd(config))

	return cmd
}
