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
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
)

const (
	CONTAINER_NAME = "xos-mock-grpc-server"
	//MOCK_DIR       = "/home/smbaker/projects/gopath/src/github.com/opencord/cordctl/mock"
)

var MockDir = os.Getenv("CORDCTL_MOCK_DIR")

func init() {
	if MockDir == "" {
		panic("CORDCTL_MOCK_DIR environment variable not set")
	}
}

// Start the mock server and wait for it to be ready
//     `data_name` is the name of the data.json to tell the mock server to use.
//     If a mock server is already running with the same data_name, it is not restarted.
func StartMockServer(data_name string) error {
	cmd_str := fmt.Sprintf("cd %s && DATA_JSON=%s docker-compose up -d", MockDir, data_name)
	cmd := exec.Command("/bin/bash", "-c", cmd_str)

	err := cmd.Run()
	if err != nil {
		return err
	}

	err = WaitForReady()
	if err != nil {
		return err
	}

	return nil
}

// Stop the mock server
func StopMockServer() error {
	cmd_str := fmt.Sprintf("cd %s && docker-compose down", MockDir)
	cmd := exec.Command("/bin/bash", "-c", cmd_str)

	err := cmd.Run()
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

// Assert that two JSON-encoded strings are equal
func AssertJSONEqual(t *testing.T, actual string, expected string) error {
	var expected_json interface{}
	err := json.Unmarshal([]byte(expected), &expected_json)
	if err != nil {
		t.Errorf("Failed to unmarshal expected json %s", expected)
		return err
	}

	var actual_json interface{}
	err = json.Unmarshal([]byte(actual), &actual_json)
	if err != nil {
		t.Errorf("Failed to unmarshal actual json %s", actual_json)
		return err
	}

	if !reflect.DeepEqual(expected_json, actual_json) {
		t.Errorf("Actual json does not match expected json\nACTUAL:\n%s\nEXPECTED:\n%s", actual, expected)
	}

	return nil
}

// Assert that the error string is what we expect
func AssertErrorEqual(t *testing.T, err error, expected string) error {
	if err == nil {
		t.Error("Expected an error, but received nil")
		return errors.New("AssertErrorEqual")
	}
	if err.Error() != expected {
		t.Errorf("Expected error `%s` but received actual error `%s`", expected, err.Error())
		return errors.New("AssertErrorEqual")
	}
	return nil
}
