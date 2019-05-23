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
	"fmt"
	"github.com/opencord/cordctl/testutils"
	"os"
	"testing"
)

func TestVersionClientOnly(t *testing.T) {
	expected := "" +
		"Client:\n" +
		" Version         unknown-version\n" +
		" Go version:     unknown-goversion\n" +
		" Git commit:     unknown-gitcommit\n" +
		" Git dirty:      unknown-gitdirty\n" +
		" Built:          unknown-buildtime\n" +
		" OS/Arch:        unknown-os/unknown-arch\n" +
		"\n"

	got := new(bytes.Buffer)
	OutputStream = got

	var options VersionOpts
	options.ClientOnly = true
	err := options.Execute([]string{})

	if err != nil {
		t.Errorf("%s: Received error %v", t.Name(), err)
		return
	}

	if got.String() != expected {
		t.Logf("RECEIVED:\n%s\n", got.String())
		t.Logf("EXPECTED:\n%s\n", expected)
		t.Errorf("%s: expected and received did not match", t.Name())
	}
}

func TestVersionClientAndServer(t *testing.T) {
	expected := "" +
		"Client:\n" +
		" Version         unknown-version\n" +
		" Go version:     unknown-goversion\n" +
		" Git commit:     unknown-gitcommit\n" +
		" Git dirty:      unknown-gitdirty\n" +
		" Built:          unknown-buildtime\n" +
		" OS/Arch:        unknown-os/unknown-arch\n" +
		"\n" +
		"Server:\n" +
		" Version         3.2.6\n" +
		" Python version: 2.7.16 (default, May  6 2019, 19:35:26)\n" +
		" Git commit:     b0df1bf6ed1698285eda6a6725c5da0c80aa4aee\n" +
		" Built:          2019-05-20T17:04:14Z\n" +
		" OS/Arch:        linux/x86_64\n" +
		"\n"

	got := new(bytes.Buffer)
	OutputStream = got

	var options VersionOpts
	err := options.Execute([]string{})

	if err != nil {
		t.Errorf("%s: Received error %v", t.Name(), err)
		return
	}

	if got.String() != expected {
		t.Logf("RECEIVED:\n%s\n", got.String())
		t.Logf("EXPECTED:\n%s\n", expected)
		t.Errorf("%s: expected and received did not match", t.Name())
	}
}

func TestMain(m *testing.M) {
	err := testutils.StartMockServer("data.json")
	if err != nil {
		fmt.Printf("Error when initializing mock server %v", err)
		os.Exit(-1)
	}
	os.Exit(m.Run())
}
