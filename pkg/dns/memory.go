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

type memoryCache struct {
	records map[string]string
}

func init() {
	Registry(&memoryCache{
		records: make(map[string]string),
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

func (m *memoryCache) Name() string {
	return "memory"
}
