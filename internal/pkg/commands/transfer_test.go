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

func TestDownload(t *testing.T) {
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

	var options TransferOpts
	options.Download.OutputAs = "json"
	options.Download.Args.LocalFileName = "/tmp/transfer.down"
	options.Download.Args.URI = "file:///tmp/transfer.down"
	err := options.Download.Execute([]string{})

	if err != nil {
		t.Errorf("%s: Received error %v", t.Name(), err)
		return
	}

	testutils.AssertJSONEqual(t, got.String(), expected)
}

func TestUpload(t *testing.T) {
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

	var options TransferOpts
	options.Upload.OutputAs = "json"
	options.Upload.Args.LocalFileName = "/tmp/transfer.up"
	options.Upload.Args.URI = "file:///tmp/transfer.up"
	options.Upload.ChunkSize = 3
	err = options.Upload.Execute([]string{})

	if err != nil {
		t.Errorf("%s: Received error %v", t.Name(), err)
		return
	}

	testutils.AssertJSONEqual(t, got.String(), expected)
}
