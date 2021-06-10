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
	"net/http"
	"testing"
	"time"

	"github.com/PagerDuty/go-pdagent/cmd/cmdutil"
	"github.com/PagerDuty/go-pdagent/test"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestEnqueue_noInput(t *testing.T) {
	defer gock.Off()

	defaultHTTPClient := &http.Client{
		Timeout: 5 * time.Minute,
	}

	realConfig := cmdutil.NewConfig()
	realConfig.HttpClient = func() (*http.Client, error) {
		return defaultHTTPClient, nil
	}

	cmd := NewEnqueueCmd(realConfig)

	gock.New(cmdutil.GetDefaults().Address).
		Post("/send").
		BodyString(`{"routing_key":"","event_action":"","payload":{"summary":"","source":"","severity":"error"}}`).
		Reply(200).
		BodyString(`{"errors":["invalid routing key"]}`)

	gock.InterceptClient(defaultHTTPClient)

	out, err := test.CaptureStdout(func() error {
		_, err := cmd.ExecuteC()
		return err
	})

	if err != nil {
		t.Errorf("error running command `enqueue`: %v", err)
	}

	assert.Contains(t, out, `{"errors":["invalid routing key"]}`)
}

func TestEnqueue_validInput(t *testing.T) {
	defer gock.Off()

	defaultHTTPClient := &http.Client{
		Timeout: 5 * time.Minute,
	}

	realConfig := cmdutil.NewConfig()
	realConfig.HttpClient = func() (*http.Client, error) {
		return defaultHTTPClient, nil
	}

	const Action = "trigger"
	const RoutingKey = "abc"
	const Source = "The Sarlacc Pit"
	const Summary = "Agent, PD Agent"

	cmd := NewEnqueueCmd(realConfig)
	cmd.SetArgs([]string{
		"-k", RoutingKey,
		"-t", Action,
		"-u", Source,
		"-d", Summary,
	})

	gock.New(cmdutil.GetDefaults().Address).
		Post("/send").
		JSON(map[string]interface{}{
			"routing_key":  RoutingKey,
			"event_action": Action,
			"payload": map[string]string{
				"summary":  Summary,
				"source":   Source,
				"severity": "error",
			},
		}).
		Reply(200).
		JSON(map[string]interface{}{"key": "xyz"})

	gock.InterceptClient(defaultHTTPClient)

	out, err := test.CaptureStdout(func() error {
		_, err := cmd.ExecuteC()
		return err
	})

	if err != nil {
		t.Errorf("error running command `enqueue`: %v", err)
	}

	assert.Contains(t, out, `{"key":"xyz"}`)
}
