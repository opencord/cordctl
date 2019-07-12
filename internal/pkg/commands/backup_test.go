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
	"io/ioutil"
	"testing"
)

func TestBackupCreate(t *testing.T) {
	// use `python -m json.tool` to pretty-print json
	expected := `[
		{
			"bytes": 6,
			"checksum": "sha256:e9c0f8b575cbfcb42ab3b78ecc87efa3b011d9a5d10b09fa4e96f240bf6a82f5",
			"chunks": 2,
			"status": "SUCCESS"
		}
	]`

	got := new(bytes.Buffer)
	OutputStream = got

	var options BackupOpts
	options.Create.OutputAs = "json"
	options.Create.Args.LocalFileName = "/tmp/transfer.down"
	options.Create.ChunkSize = 3
	err := options.Create.Execute([]string{})

	if err != nil {
		t.Errorf("%s: Received error %v", t.Name(), err)
		return
	}

	testutils.AssertJSONEqual(t, got.String(), expected)
}

// Mock the CreateURI function in the Restore code to use file:///tmp/transfer.up
func CreateURI() (string, string) {
	remote_name := "transfer.up"
	uri := "file:///tmp/" + remote_name
	return remote_name, uri
}

func TestBackupRestore(t *testing.T) {
	// use `python -m json.tool` to pretty-print json
	expected := `[
		{
			"bytes": 6,
			"checksum": "sha256:e9c0f8b575cbfcb42ab3b78ecc87efa3b011d9a5d10b09fa4e96f240bf6a82f5",
			"chunks": 2,
			"status": "SUCCESS"
		}
	]`

	err := ioutil.WriteFile("/tmp/transfer.up", []byte("ABCDEF"), 0644)

	got := new(bytes.Buffer)
	OutputStream = got

	var options BackupRestore
	options.OutputAs = "json"
	options.Args.LocalFileName = "/tmp/transfer.up"
	options.ChunkSize = 3
	options.CreateURIFunc = CreateURI
	err = options.Execute([]string{})

	if err != nil {
		t.Errorf("%s: Received error %v", t.Name(), err)
		return
	}

	testutils.AssertJSONEqual(t, got.String(), expected)
}
