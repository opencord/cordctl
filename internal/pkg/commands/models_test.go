/*
 * Portions copyright 2019-present Open Networking Foundation
 * Original copyright 2019-present Ciena Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the"github.com/stretchr/testify/assert" "License");
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
	corderrors "github.com/opencord/cordctl/internal/pkg/error"
	"github.com/opencord/cordctl/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
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

func TestModelListDirty(t *testing.T) {
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
	options.List.State = "dirty"
	err := options.List.Execute([]string{})

	if err != nil {
		t.Errorf("%s: Received error %v", t.Name(), err)
		return
	}

	testutils.AssertJSONEqual(t, got.String(), expected)
}

func TestModelUpdate(t *testing.T) {
	expected := `[{"id":1, "message":"Updated"}]`

	got := new(bytes.Buffer)
	OutputStream = got

	var options ModelOpts
	options.Update.Args.ModelName = "Slice"
	options.Update.OutputAs = "json"
	options.Update.IDArgs.ID = []int32{1}
	options.Update.SetFields = "name=mockslice1_newname"
	err := options.Update.Execute([]string{})

	if err != nil {
		t.Errorf("%s: Received error %v", t.Name(), err)
		return
	}

	testutils.AssertJSONEqual(t, got.String(), expected)
}

func TestModelUpdateUsingFilter(t *testing.T) {
	expected := `[{"id":1, "message":"Updated"}]`

	got := new(bytes.Buffer)
	OutputStream = got

	var options ModelOpts
	options.Update.Args.ModelName = "Slice"
	options.Update.OutputAs = "json"
	options.Update.Filter = "id=1"
	options.Update.SetFields = "name=mockslice1_newname"
	err := options.Update.Execute([]string{})

	if err != nil {
		t.Errorf("%s: Received error %v", t.Name(), err)
		return
	}

	testutils.AssertJSONEqual(t, got.String(), expected)
}

func TestModelUpdateNoExist(t *testing.T) {
	got := new(bytes.Buffer)
	OutputStream = got

	var options ModelOpts
	options.Update.Args.ModelName = "Slice"
	options.Update.OutputAs = "json"
	options.Update.IDArgs.ID = []int32{77}
	options.Update.SetFields = "name=mockslice1_newname"
	err := options.Update.Execute([]string{})

	_, matched := err.(*corderrors.ModelNotFoundError)
	assert.True(t, matched)
}

func TestModelUpdateUsingFilterNoExist(t *testing.T) {
	got := new(bytes.Buffer)
	OutputStream = got

	var options ModelOpts
	options.Update.Args.ModelName = "Slice"
	options.Update.OutputAs = "json"
	options.Update.Filter = "id=77"
	options.Update.SetFields = "name=mockslice1_newname"
	err := options.Update.Execute([]string{})

	_, matched := err.(*corderrors.NoMatchError)
	assert.True(t, matched)
}

func TestModelCreate(t *testing.T) {
	expected := `[{"id":3, "message":"Created"}]`

	got := new(bytes.Buffer)
	OutputStream = got

	var options ModelOpts
	options.Create.Args.ModelName = "Slice"
	options.Create.OutputAs = "json"
	options.Create.SetFields = "name=mockslice3,site_id=1"
	err := options.Create.Execute([]string{})

	if err != nil {
		t.Errorf("%s: Received error %v", t.Name(), err)
		return
	}

	testutils.AssertJSONEqual(t, got.String(), expected)
}

func TestModelDelete(t *testing.T) {
	expected := `[{"id":1, "message":"Deleted"}]`

	got := new(bytes.Buffer)
	OutputStream = got

	var options ModelOpts
	options.Delete.Args.ModelName = "Slice"
	options.Delete.OutputAs = "json"
	options.Delete.IDArgs.ID = []int32{1}
	err := options.Delete.Execute([]string{})

	if err != nil {
		t.Errorf("%s: Received error %v", t.Name(), err)
		return
	}

	testutils.AssertJSONEqual(t, got.String(), expected)
}

func TestModelDeleteUsingFilter(t *testing.T) {
	expected := `[{"id":1, "message":"Deleted"}]`

	got := new(bytes.Buffer)
	OutputStream = got

	var options ModelOpts
	options.Delete.Args.ModelName = "Slice"
	options.Delete.OutputAs = "json"
	options.Delete.Filter = "id=1"
	err := options.Delete.Execute([]string{})

	if err != nil {
		t.Errorf("%s: Received error %v", t.Name(), err)
		return
	}

	testutils.AssertJSONEqual(t, got.String(), expected)
}

func TestModelDeleteNoExist(t *testing.T) {
	expected := `[{"id":77, "message":"Not Found [on model Slice <id=77>]"}]`

	got := new(bytes.Buffer)
	OutputStream = got

	var options ModelOpts
	options.Delete.Args.ModelName = "Slice"
	options.Delete.OutputAs = "json"
	options.Delete.IDArgs.ID = []int32{77}
	err := options.Delete.Execute([]string{})

	if err != nil {
		t.Errorf("%s: Received error %v", t.Name(), err)
		return
	}

	testutils.AssertJSONEqual(t, got.String(), expected)
}

func TestModelDeleteFilterNoExist(t *testing.T) {
	got := new(bytes.Buffer)
	OutputStream = got

	var options ModelOpts
	options.Delete.Args.ModelName = "Slice"
	options.Delete.OutputAs = "json"
	options.Delete.Filter = "id=77"
	err := options.Delete.Execute([]string{})

	_, matched := err.(*corderrors.NoMatchError)
	assert.True(t, matched)
}

func TestModelSync(t *testing.T) {
	expected := `[{"id":1, "message":"Enacted"}]`

	got := new(bytes.Buffer)
	OutputStream = got

	var options ModelOpts
	options.Sync.Args.ModelName = "Slice"
	options.Sync.OutputAs = "json"
	options.Sync.IDArgs.ID = []int32{1}
	options.Sync.SyncTimeout = 5 * time.Second
	err := options.Sync.Execute([]string{})

	if err != nil {
		t.Errorf("%s: Received error %v", t.Name(), err)
		return
	}

	testutils.AssertJSONEqual(t, got.String(), expected)
}

func TestModelSyncTimeout(t *testing.T) {
	expected := `[{"id":2, "message":"context deadline exceeded"}]`

	got := new(bytes.Buffer)
	OutputStream = got

	var options ModelOpts
	options.Sync.Args.ModelName = "Slice"
	options.Sync.OutputAs = "json"
	options.Sync.IDArgs.ID = []int32{2}
	options.Sync.SyncTimeout = 5 * time.Second
	err := options.Sync.Execute([]string{})

	if err != nil {
		t.Errorf("%s: Received error %v", t.Name(), err)
		return
	}

	testutils.AssertJSONEqual(t, got.String(), expected)
}

func TestModelSetDirty(t *testing.T) {
	expected := `[{"id":1, "message":"Dirtied"}]`

	got := new(bytes.Buffer)
	OutputStream = got

	var options ModelOpts
	options.SetDirty.Args.ModelName = "Slice"
	options.SetDirty.OutputAs = "json"
	options.SetDirty.IDArgs.ID = []int32{1}
	err := options.SetDirty.Execute([]string{})

	if err != nil {
		t.Errorf("%s: Received error %v", t.Name(), err)
		return
	}

	testutils.AssertJSONEqual(t, got.String(), expected)
}
