/*
 * Portions copyright 2019-present Open Networking Foundation
 * Original copyright 2019-present Ciena Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package commands

import (
	"bytes"
	"github.com/opencord/cordctl/pkg/testutils"
	"testing"
)

func TestServicesList(t *testing.T) {
	// use `python -m json.tool` to pretty-print json
	expected := `[
		{
			"name": "onos",
			"state": "present",
			"version": "2.1.1-dev"
		},
		{
			"name": "kubernetes",
			"state": "present",
			"version": "1.2.1"
		}
	]`

	got := new(bytes.Buffer)
	OutputStream = got

	var options ServiceOpts
	options.List.OutputAs = "json"
	err := options.List.Execute([]string{})

	if err != nil {
		t.Errorf("%s: Received error %v", t.Name(), err)
		return
	}

	testutils.AssertJSONEqual(t, got.String(), expected)
}

func TestServicesListTable(t *testing.T) {
	// We'll use the ServicesList command to be our end-to-end test for
	// table formatted commands, as the output is relatively simple.
	expected := `NAME          VERSION      STATE
onos          2.1.1-dev    present
kubernetes    1.2.1        present
`

	got := new(bytes.Buffer)
	OutputStream = got

	var options ServiceOpts
	err := options.List.Execute([]string{})

	if err != nil {
		t.Errorf("%s: Received error %v", t.Name(), err)
		return
	}

	testutils.AssertStringEqual(t, got.String(), expected)
}

func TestServicesListYaml(t *testing.T) {
	// We'll use the ServicesList command to be our end-to-end test for
	// yaml formatted commands, as the output is relatively simple.
	expected := `- name: onos
  version: 2.1.1-dev
  state: present
- name: kubernetes
  version: 1.2.1
  state: present
`

	got := new(bytes.Buffer)
	OutputStream = got

	var options ServiceOpts
	options.List.OutputAs = "yaml"
	err := options.List.Execute([]string{})

	if err != nil {
		t.Errorf("%s: Received error %v", t.Name(), err)
		return
	}

	testutils.AssertStringEqual(t, got.String(), expected)
}
