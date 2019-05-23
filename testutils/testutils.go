/*
 * Copyright 2019-present Open Networking Foundation
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
package testutils

import (
	"fmt"
	"os/exec"
	"strings"
)

const (
	CONTAINER_NAME = "xos-mock-grpc-server"
)

// Start the mock server and wait for it to be ready
func StartMockServer(data_name string) error {
	mock_dir := "/home/smbaker/projects/gopath/src/github.com/opencord/cordctl/mock"
	cmd_str := fmt.Sprintf("cd %s && DATA_JSON=%s docker-compose up -d", mock_dir, data_name)
	cmd := exec.Command("/bin/bash", "-c", cmd_str)

	out, err := cmd.Output()
	if err != nil {
		return err
	}

	_ = out

	err = WaitForReady()
	if err != nil {
		return err
	}

	return nil
}

// Wait for the mock server to be ready
func WaitForReady() error {
	for {
		ready, err := IsReady()
		if err != nil {
			return err
		}
		if ready {
			return nil
		}
	}
}

// Return true if the mock server is ready
func IsReady() (bool, error) {
	cmd := exec.Command("docker", "logs", CONTAINER_NAME)
	out, err := cmd.Output()
	if err != nil {
		return false, err
	}

	return strings.Contains(string(out), "Listening for requests"), nil
}
