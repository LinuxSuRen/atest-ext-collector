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

type DNSCache interface {
	LookupIP(string) string
	Init(map[string]string)
	Put(domain, ip string)
	Remove(string)
	Data() map[string]string
	Size() int
	AddBlackDomain(string)
	RemoveBlackDomain(string)
	ListBlackDomains() []string
	IsBlackDomain(string) bool
	GetWildcardCache() DNSCache
	Name() string
}

type Server interface {
	Start() error
	Stop() error
}
