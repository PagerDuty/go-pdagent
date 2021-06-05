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
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/PagerDuty/go-pdagent/test"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestNagiosEnqueue_missingRequiredFlags(t *testing.T) {
	realConfig := New()

	cmd := NewNagiosEnqueueCmd(realConfig)

	_, err := cmd.ExecuteC()

	expectedErr := errors.New("required flag(s) \"notification-type\", \"routing-key\", \"severity\", \"source-type\" not set")
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestNagiosEnqueue_invalidNotificationType(t *testing.T) {
	realConfig := New()

	const notificationType = "trigger"
	const routingKey = "abc"
	const sourceType = "host"
	const severity = "info"

	cmd := NewNagiosEnqueueCmd(realConfig)
	cmd.SetArgs([]string{
		"-k", routingKey,
		"-t", notificationType,
		"-u", sourceType,
		"-e", severity,
	})

	_, err := cmd.ExecuteC()

	assert.Error(t, err)
	assert.Equal(t, errNotificationType, err)
}

func TestNagiosEnqueue_invalidSourceType(t *testing.T) {
	realConfig := New()

	const notificationType = "PROBLEM"
	const routingKey = "abc"
	const sourceType = "invalidSourceType"
	const severity = "warning"

	cmd := NewNagiosEnqueueCmd(realConfig)
	cmd.SetArgs([]string{
		"-k", routingKey,
		"-t", notificationType,
		"-u", sourceType,
		"-e", severity,
	})

	_, err := cmd.ExecuteC()

	assert.Error(t, err)
	assert.Equal(t, errSourceType, err)
}

func TestNagiosEnqueue_invalidSeverity(t *testing.T) {
	realConfig := New()

	const notificationType = "ACKNOWLEDGEMENT"
	const routingKey = "abc"
	const sourceType = "service"
	const severity = "invalidSeverity"

	cmd := NewNagiosEnqueueCmd(realConfig)
	cmd.SetArgs([]string{
		"-k", routingKey,
		"-t", notificationType,
		"-u", sourceType,
		"-e", severity,
	})

	_, err := cmd.ExecuteC()

	assert.Error(t, err)
	assert.Equal(t, errSeverity, err)
}

func TestNagiosEnqueue_invalidServiceCustomDetails(t *testing.T) {
	realConfig := New()

	const notificationType = "RECOVERY"
	const routingKey = "abc"
	const sourceType = "service"
	const severity = "critical"

	cmd := NewNagiosEnqueueCmd(realConfig)
	cmd.SetArgs([]string{
		"-k", routingKey,
		"-t", notificationType,
		"-u", sourceType,
		"-e", severity,
	})

	_, err := cmd.ExecuteC()

	expectedErr := errors.New("The HOSTNAME field must be set for source-type \"service\" using the -f flag")
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)

	cmd.SetArgs([]string{
		"-k", routingKey,
		"-t", notificationType,
		"-u", sourceType,
		"-e", severity,
		"-f", "HOSTNAME=computer.network",
	})

	_, err = cmd.ExecuteC()

	expectedErr = errors.New("The SERVICEDESC field must be set for source-type \"service\" using the -f flag")
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)

	cmd.SetArgs([]string{
		"-k", routingKey,
		"-t", notificationType,
		"-u", sourceType,
		"-e", severity,
		"-f", "HOSTNAME=computer.network",
		"-f", "SERVICEDESC=a service",
	})

	_, err = cmd.ExecuteC()

	expectedErr = errors.New("The SERVICESTATE field must be set for source-type \"service\" using the -f flag")
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestNagiosEnqueue_invalidHostCustomDetails(t *testing.T) {
	realConfig := New()

	const notificationType = "RECOVERY"
	const routingKey = "abc"
	const sourceType = "host"
	const severity = "error"

	cmd := NewNagiosEnqueueCmd(realConfig)
	cmd.SetArgs([]string{
		"-k", routingKey,
		"-t", notificationType,
		"-u", sourceType,
		"-e", severity,
	})

	_, err := cmd.ExecuteC()

	expectedErr := errors.New("The HOSTNAME field must be set for source-type \"host\" using the -f flag")
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)

	cmd.SetArgs([]string{
		"-k", routingKey,
		"-t", notificationType,
		"-u", sourceType,
		"-e", severity,
		"-f", "HOSTNAME=computer.network",
	})

	_, err = cmd.ExecuteC()

	expectedErr = errors.New("The HOSTSTATE field must be set for source-type \"host\" using the -f flag")
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestNagiosEnqueue_validSourceHostInput(t *testing.T) {
	defer gock.Off()

	defaultHTTPClient := &http.Client{
		Timeout: 5 * time.Minute,
	}

	realConfig := New()
	realConfig.HttpClient = func() (*http.Client, error) {
		return defaultHTTPClient, nil
	}

	const notificationType = "PROBLEM"
	const routingKey = "xyz"
	const sourceType = "host"
	const severity = "warning"
	const hostname = "computer.network"
	const hoststate = "down"

	cmd := NewNagiosEnqueueCmd(realConfig)
	cmd.SetArgs([]string{
		"-k", routingKey,
		"-t", notificationType,
		"-u", sourceType,
		"-e", severity,
		"-f", fmt.Sprintf("HOSTNAME=%v", hostname),
		"-f", fmt.Sprintf("HOSTSTATE=%v", hoststate),
	})

	gock.New(getDefaults().Address).
		Post("/send").
		JSON(map[string]interface{}{
			"routing_key":  routingKey,
			"event_action": nagiosToPagerDutyEventType[notificationType],
			"dedup_key":    fmt.Sprintf("event_source=%v;host_name=%v", sourceType, hostname),
			"payload": map[string]interface{}{
				"summary":  fmt.Sprintf("HOSTNAME=%v; HOSTSTATE=%v", hostname, hoststate),
				"source":   hostname,
				"severity": severity,
				"custom_details": map[string]string{
					"pd_nagios_object": sourceType,
					"HOSTNAME":         hostname,
					"HOSTSTATE":        hoststate,
				},
			},
		}).
		Reply(200).
		JSON(map[string]interface{}{"key": routingKey})

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

func TestNagiosEnqueue_validSourceServiceInput(t *testing.T) {
	defer gock.Off()

	defaultHTTPClient := &http.Client{
		Timeout: 5 * time.Minute,
	}

	realConfig := New()
	realConfig.HttpClient = func() (*http.Client, error) {
		return defaultHTTPClient, nil
	}

	const notificationType = "PROBLEM"
	const routingKey = "xyz"
	const sourceType = "service"
	const severity = "warning"
	const hostname = "computer.network"
	const serviceDesc = "some service desc"
	const serviceState = "down"

	cmd := NewNagiosEnqueueCmd(realConfig)
	cmd.SetArgs([]string{
		"-k", routingKey,
		"-t", notificationType,
		"-u", sourceType,
		"-e", severity,
		"-f", fmt.Sprintf("HOSTNAME=%v", hostname),
		"-f", fmt.Sprintf("SERVICEDESC=%v", serviceDesc),
		"-f", fmt.Sprintf("SERVICESTATE=%v", serviceState),
	})

	gock.New(getDefaults().Address).
		Post("/send").
		JSON(map[string]interface{}{
			"routing_key":  routingKey,
			"event_action": nagiosToPagerDutyEventType[notificationType],
			"dedup_key":    fmt.Sprintf("event_source=%v;host_name=%v;service_desc=%v", sourceType, hostname, serviceDesc),
			"payload": map[string]interface{}{
				"summary":  fmt.Sprintf("HOSTNAME=%v; SERVICEDESC=%v; SERVICESTATE=%v", hostname, serviceDesc, serviceState),
				"source":   hostname,
				"severity": severity,
				"custom_details": map[string]string{
					"pd_nagios_object": sourceType,
					"HOSTNAME":         hostname,
					"SERVICEDESC":      serviceDesc,
					"SERVICESTATE":     serviceState,
				},
			},
		}).
		Reply(200).
		JSON(map[string]interface{}{"key": routingKey})

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

func TestNagiosEnqueue_userProvidedDedupKey(t *testing.T) {
	defer gock.Off()

	defaultHTTPClient := &http.Client{
		Timeout: 5 * time.Minute,
	}

	realConfig := New()
	realConfig.HttpClient = func() (*http.Client, error) {
		return defaultHTTPClient, nil
	}

	const notificationType = "PROBLEM"
	const routingKey = "xyz"
	const sourceType = "service"
	const severity = "warning"
	const hostname = "computer.network"
	const serviceDesc = "some service desc"
	const serviceState = "down"
	const dedupKey = "someDedupKey"

	cmd := NewNagiosEnqueueCmd(realConfig)
	cmd.SetArgs([]string{
		"-k", routingKey,
		"-t", notificationType,
		"-u", sourceType,
		"-e", severity,
		"-y", dedupKey,
		"-f", fmt.Sprintf("HOSTNAME=%v", hostname),
		"-f", fmt.Sprintf("SERVICEDESC=%v", serviceDesc),
		"-f", fmt.Sprintf("SERVICESTATE=%v", serviceState),
	})

	gock.New(getDefaults().Address).
		Post("/send").
		JSON(map[string]interface{}{
			"routing_key":  routingKey,
			"event_action": nagiosToPagerDutyEventType[notificationType],
			"dedup_key":    dedupKey,
			"payload": map[string]interface{}{
				"summary":  fmt.Sprintf("HOSTNAME=%v; SERVICEDESC=%v; SERVICESTATE=%v", hostname, serviceDesc, serviceState),
				"source":   hostname,
				"severity": severity,
				"custom_details": map[string]string{
					"pd_nagios_object": sourceType,
					"HOSTNAME":         hostname,
					"SERVICEDESC":      serviceDesc,
					"SERVICESTATE":     serviceState,
				},
			},
		}).
		Reply(200).
		JSON(map[string]interface{}{"key": routingKey})

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
