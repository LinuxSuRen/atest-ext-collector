/*
Copyright 2024 LinuxSuRen.

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
	"fmt"
	"regexp"
	"strings"
)

type memoryCache struct {
	records       map[string]string
	black         []string
	wildcardCache DNSCache
}

type memoryWildcardCache struct {
	*memoryCache
}

func init() {
	Registry(&memoryCache{
		records: make(map[string]string),
		black:   []string{},
		wildcardCache: &memoryWildcardCache{
			memoryCache: &memoryCache{
				records: make(map[string]string),
				black:   []string{},
			},
		},
	})
}

func (m *memoryCache) LookupIP(domain string) (ip string) {
	ip = m.records[domain]
	return
}

func (m *memoryCache) Init(init map[string]string) {
	m.records = init
}

func (m *memoryCache) Put(domain, ip string) {
	domain = strings.TrimSpace(domain)
	if domain == "" {
		return
	}
	m.records[domain] = ip
}
func (m *memoryCache) Remove(domain string) {
	delete(m.records, domain)
}

func (m *memoryCache) Data() (data map[string]string) {
	data = make(map[string]string)
	for k, v := range m.records {
		data[k] = v
	}
	return
}

func (m *memoryCache) Size() int {
	return len(m.records)
}

func (m *memoryCache) AddBlackDomain(domain string) {
	domain = strings.TrimSpace(domain)
	if domain == "" {
		return
	}
	m.black = append(m.black, domain)
	return
}
func (m *memoryCache) RemoveBlackDomain(domain string) {
	for i, item := range m.black {
		if item == domain {
			m.black = append(m.black[:i], m.black[i+1:]...)
			break
		}
	}
}
func (m *memoryCache) ListBlackDomains() (items []string) {
	items = make([]string, len(m.black))
	copy(items, m.black)
	return
}

func (m *memoryCache) IsBlackDomain(domain string) bool {
	for _, item := range m.black {
		if item == domain {
			return true
		}
	}
	return false
}

func (m *memoryCache) GetWildcardCache() DNSCache {
	return m.wildcardCache
}

func (m *memoryCache) Name() string {
	return "memory"
}

func (m *memoryWildcardCache) LookupIP(domain string) string {
	fmt.Println("looking", domain, "from wildcard")
	for pattern, ip := range m.records {
		matched, _ := regexp.MatchString(pattern, domain)
		if matched {
			return ip
		}
	}
	return ""
}

func (m *memoryWildcardCache) Init(init map[string]string) {
	m.records = init
}

func (m *memoryWildcardCache) Name() string {
	return "memory_wildcard"
}
