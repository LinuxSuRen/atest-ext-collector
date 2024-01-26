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

package cmd

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/linuxsuren/atest-ext-collector/pkg"
	"github.com/linuxsuren/atest-ext-collector/pkg/filter"
	"github.com/stretchr/testify/assert"
)

func TestNewRootCmd(t *testing.T) {
	c := createCollectorCmd()
	assert.NotNil(t, c)
	assert.Equal(t, "collector", c.Use)
}

func TestResponseFilter(t *testing.T) {
	targetURL, err := url.Parse("http://foo.com/api/v1")
	assert.NoError(t, err)

	resp := &http.Response{
		Header: http.Header{
			"Content-Type": []string{"application/json; charset=utf-8"},
		},
		Request: &http.Request{
			URL: targetURL,
		},
		Body: io.NopCloser(bytes.NewBuffer([]byte("hello"))),
	}
	emptyResp := &http.Response{}

	filter := &responseFilter{
		urlFilter: &filter.URLPathFilter{
			PathPrefix: []string{"/api/v1"},
		},
		collects: pkg.NewCollects(),
		ctx:      context.Background(),
	}
	filter.filter(emptyResp, nil)
	filter.filter(resp, nil)
}
