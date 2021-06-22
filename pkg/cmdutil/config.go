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
	"net/http"
	"time"

	"github.com/PagerDuty/go-pdagent/pkg/client"
	"github.com/spf13/viper"
)

type Config struct {
	HttpClient func() (*http.Client, error)
	Client     func() (*client.Client, error)
}

func NewConfig() *Config {
	httpClientFunc := func() (*http.Client, error) {
		client := &http.Client{
			Transport: http.DefaultTransport,
			Timeout:   5 * time.Second,
		}
		return client, nil
	}

	return &Config{
		HttpClient: httpClientFunc,
		Client: func() (*client.Client, error) {
			httpClient, _ := httpClientFunc()
			c := client.NewClient(httpClient, viper.GetString("address"), viper.GetString("secret"))
			return c, nil
		},
	}
}
