/*
Copyright 2025 LinuxSuRen.

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

package dns_test

import (
	"github.com/linuxsuren/atest-ext-collector/pkg/dns"
	"github.com/stretchr/testify/assert"
	"testing"
)
import _ "embed"

func TestParse(t *testing.T) {
	defaultConfig := &dns.DNSConfig{
		Wildcard: map[string]string{
			"*.def.com": "0.0.0.2",
		},
		Simple: map[string]string{
			"abc.com": "0.0.0.1",
		},
		Black:         []string{"ghi.com"},
		WildcardBlack: []string{"*.jkl.com"},
		Upstream:      "8.8.8.8",
		Port:          53,
	}

	config, err := dns.ParseFromBuffer(configYaml)
	assert.NoError(t, err)
	assert.Equal(t, defaultConfig, config)

	config, err = dns.ParseFromFile("testdata/config.yaml")
	assert.NoError(t, err)
	assert.Equal(t, defaultConfig, config)
}

var (
	//go:embed testdata/config.yaml
	configYaml []byte
)
