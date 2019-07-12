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
	"fmt"
	"github.com/opencord/cordctl/pkg/testutils"
	"os"
	"testing"
)

// This TestMain is global to all tests in the `commands` package

func TestMain(m *testing.M) {
	err := testutils.StartMockServer("data.json")
	if err != nil {
		fmt.Printf("Error when initializing mock server %v", err)
		os.Exit(-1)
	}
	os.Exit(m.Run())
}
