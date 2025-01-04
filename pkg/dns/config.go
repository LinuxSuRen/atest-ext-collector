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

package dns

import (
	"gopkg.in/yaml.v3"
	"os"
)

type DNSConfig struct {
	Wildcard      map[string]string `yaml:"wildcard"`
	Simple        map[string]string `yaml:"simple"`
	Black         []string          `yaml:"black"`
	WildcardBlack []string          `yaml:"wildcard_black"`
	Upstream      string            `yaml:"upstream"`
	Port          int               `yaml:"port"`
}

func ParseFromFile(file string) (config *DNSConfig, err error) {
	var data []byte
	if data, err = os.ReadFile(file); err == nil {
		config, err = ParseFromBuffer(data)
	}
	return
}

func ParseFromBuffer(buffer []byte) (config *DNSConfig, err error) {
	config = &DNSConfig{}
	err = yaml.Unmarshal(buffer, &config)
	return
}
