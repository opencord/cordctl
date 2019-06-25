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
package error

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
)

func TestGenericError(t *testing.T) {
	var err error

	err = fmt.Errorf("Some error")

	// Type conversion from `error` to ChecksumMismatchError should fail
	_, ok := err.(ChecksumMismatchError)
	assert.False(t, ok)

	// Type conversion from `error` to CordCtlError should fail
	_, ok = err.(CordCtlError)
	assert.False(t, ok)
}

func TestChecksumMismatchError(t *testing.T) {
	var err error

	err = WithStackTrace(&ChecksumMismatchError{Actual: "123", Expected: "456"})

	// Check that the Error() function returns the right text
	assert.Equal(t, err.Error(), "checksum mismatch (actual=456, expected=123)")

	// Type conversion from `error` to ChecksumMismatchError should succeed
	_, ok := err.(*ChecksumMismatchError)
	assert.True(t, ok)

	// Type switch is another way of doing the same
	switch err.(type) {
	case *ChecksumMismatchError:
		// do nothing
	case CordCtlError:
		assert.Fail(t, "Should have used the ChecksumMismatchError case instead")
	default:
		assert.Fail(t, "Wrong part of switch statement was called")
	}

	// Type conversion from `error` to CordCtlError should succeed
	cce, ok := err.(CordCtlError)
	assert.True(t, ok)

	// ShouldDumpStack() returned from a ChecksumMismatchError should be false
	assert.False(t, cce.ShouldDumpStack())
}

func TestUnknownModelTypeError(t *testing.T) {
	var err error

	err = WithStackTrace(&UnknownModelTypeError{Name: "foo"})

	// Check that the Error() function returns the right text
	assert.Equal(t, err.Error(), "Model foo does not exist. Use `cordctl modeltype list` to get a list of available models")
}

func TestRpcErrorToCordError(t *testing.T) {
	// InternalError
	err := status.Error(codes.Unknown, "A fake Unknown error")

	cordErr := RpcErrorToCordError(err)

	_, ok := cordErr.(*InternalError)
	assert.True(t, ok)
	assert.Equal(t, cordErr.Error(), "Internal Error: A fake Unknown error")

	// NotFound
	err = status.Error(codes.NotFound, "A fake not found error")

	cordErr = RpcErrorToCordError(err)

	_, ok = cordErr.(*ModelNotFoundError)
	assert.True(t, ok)
	assert.Equal(t, cordErr.Error(), "Not Found")

	// PermissionDeniedError
	err = status.Error(codes.PermissionDenied, "A fake Permission error")

	cordErr = RpcErrorToCordError(err)

	_, ok = cordErr.(*PermissionDeniedError)
	assert.True(t, ok)
	assert.Equal(t, cordErr.Error(), "Permission Denied. Please verify username and password are correct")
}

func TestRpcErrorWithModelNameToCordError(t *testing.T) {
	// InternalError
	err := status.Error(codes.Unknown, "A fake Unknown error")

	cordErr := RpcErrorWithModelNameToCordError(err, "Foo")

	_, ok := cordErr.(*InternalError)
	assert.True(t, ok)
	assert.Equal(t, cordErr.Error(), "Internal Error [on model Foo]: A fake Unknown error")
}

func TestRpcErrorWithIdToCordError(t *testing.T) {
	// InternalError
	err := status.Error(codes.Unknown, "A fake Unknown error")

	cordErr := RpcErrorWithIdToCordError(err, "Foo", 7)

	_, ok := cordErr.(*InternalError)
	assert.True(t, ok)
	assert.Equal(t, cordErr.Error(), "Internal Error [on model Foo <id=7>]: A fake Unknown error")
}

func TestRpcErrorWithQueriesToCordError(t *testing.T) {
	// InternalError
	err := status.Error(codes.Unknown, "A fake Unknown error")

	cordErr := RpcErrorWithQueriesToCordError(err, "Foo", map[string]string{"id": "=3"})

	_, ok := cordErr.(*InternalError)
	assert.True(t, ok)
	assert.Equal(t, cordErr.Error(), "Internal Error [on model Foo <id=3>]: A fake Unknown error")
}

func TestStackTrace(t *testing.T) {
	var err error

	err = WithStackTrace(&UnknownModelTypeError{Name: "foo"})

	// goexit occurs near the end of the stack trace
	assert.Contains(t, err.(CordCtlError).Stack(), "goexit")
}
