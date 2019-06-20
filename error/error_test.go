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
	"testing"
)

func init() {
	SetPrefix("cordctl")
}

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

	//assert.Equal(t, err.(*ChecksumMismatchError).Stack(), "foo")

	// Check that the Error() function returns the right text
	assert.Equal(t, err.Error(), "cordctl: checksum mismatch (actual=456, expected=123)")

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

	_ = err
}
