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
	"github.com/opencord/cordctl/testutils"
	"testing"
)

func TestModelList(t *testing.T) {
	// use `python -m json.tool` to pretty-print json
	expected := `[
		{
			"controller_kind": "",
			"controller_replica_count": 0,
			"creator_id": 0,
			"default_flavor_id": 0,
			"default_image_id": 0,
			"default_isolation": "",
			"default_node_id": 0,
			"description": "",
			"enabled": false,
			"exposed_ports": "",
			"id": 1,
			"max_instances": 0,
			"mount_data_sets": "",
			"name": "mockslice1",
			"network": "",
			"principal_id": 0,
			"service_id": 0,
			"site_id": 1,
			"trust_domain_id": 0
		},
		{
			"controller_kind": "",
			"controller_replica_count": 0,
			"creator_id": 0,
			"default_flavor_id": 0,
			"default_image_id": 0,
			"default_isolation": "",
			"default_node_id": 0,
			"description": "",
			"enabled": false,
			"exposed_ports": "",
			"id": 2,
			"max_instances": 0,
			"mount_data_sets": "",
			"name": "mockslice2",
			"network": "",
			"principal_id": 0,
			"service_id": 0,
			"site_id": 1,
			"trust_domain_id": 0
		}
	]`

	got := new(bytes.Buffer)
	OutputStream = got

	var options ModelOpts
	options.List.Args.ModelName = "Slice"
	options.List.OutputAs = "json"
	err := options.List.Execute([]string{})

	if err != nil {
		t.Errorf("%s: Received error %v", t.Name(), err)
		return
	}

	testutils.AssertJSONEqual(t, got.String(), expected)
}

func TestModelListFilterID(t *testing.T) {
	// use `python -m json.tool` to pretty-print json
	expected := `[
		{
			"controller_kind": "",
			"controller_replica_count": 0,
			"creator_id": 0,
			"default_flavor_id": 0,
			"default_image_id": 0,
			"default_isolation": "",
			"default_node_id": 0,
			"description": "",
			"enabled": false,
			"exposed_ports": "",
			"id": 1,
			"max_instances": 0,
			"mount_data_sets": "",
			"name": "mockslice1",
			"network": "",
			"principal_id": 0,
			"service_id": 0,
			"site_id": 1,
			"trust_domain_id": 0
		}
	]`

	got := new(bytes.Buffer)
	OutputStream = got

	var options ModelOpts
	options.List.Args.ModelName = "Slice"
	options.List.OutputAs = "json"
	options.List.Filter = "id=1"
	err := options.List.Execute([]string{})

	if err != nil {
		t.Errorf("%s: Received error %v", t.Name(), err)
		return
	}

	testutils.AssertJSONEqual(t, got.String(), expected)
}

func TestModelListFilterName(t *testing.T) {
	// use `python -m json.tool` to pretty-print json
	expected := `[
		{
			"controller_kind": "",
			"controller_replica_count": 0,
			"creator_id": 0,
			"default_flavor_id": 0,
			"default_image_id": 0,
			"default_isolation": "",
			"default_node_id": 0,
			"description": "",
			"enabled": false,
			"exposed_ports": "",
			"id": 2,
			"max_instances": 0,
			"mount_data_sets": "",
			"name": "mockslice2",
			"network": "",
			"principal_id": 0,
			"service_id": 0,
			"site_id": 1,
			"trust_domain_id": 0
		}
	]`

	got := new(bytes.Buffer)
	OutputStream = got

	var options ModelOpts
	options.List.Args.ModelName = "Slice"
	options.List.OutputAs = "json"
	options.List.Filter = "name=mockslice2"
	err := options.List.Execute([]string{})

	if err != nil {
		t.Errorf("%s: Received error %v", t.Name(), err)
		return
	}

	testutils.AssertJSONEqual(t, got.String(), expected)
}
