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
	"strings"
	"testing"

	"github.com/PagerDuty/go-pdagent/test"
)

func TestVersionCommand(t *testing.T) {
	cmd := NewVersionCmd()

	out, err := test.CaptureStdout(func() error {
		_, err := cmd.ExecuteC()
		return err
	})

	if err != nil {
		t.Error("Did not expect error")
	}

	if !strings.Contains(out, "Version:") {
		t.Error("Expected 'version' output")
	}

	if !strings.Contains(out, "Build date:") {
		t.Error("Expected 'build date' output")
	}

	if !strings.Contains(out, "Build commit:") {
		t.Error("Expected 'build commit' output")
	}
}
