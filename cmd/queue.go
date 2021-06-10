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
	"github.com/PagerDuty/go-pdagent/cmd/cmdutil"
	"github.com/spf13/cobra"
)

func NewQueueCmd(config *cmdutil.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "queue",
		Short: "Access the daemon's event queue.",
	}

	cmd.AddCommand(NewQueueRetryCmd(config))
	cmd.AddCommand(NewQueueStatusCmd(config))

	return cmd
}
