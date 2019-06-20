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

/*  Cordctl error classes

	The basic idea is to throw specific error classes, so it's easier to test for them rather than doing string
	comparisons or other ad hoc mechanisms for determining the type of error. This decouples the human
	readable text of an error from programmatic testing of error type.

	We differentiate between errors that we want to generate brief output, such as for example a
	user mistyping a model name, versus errors that we want to generate additional context. This prevents
	overwhelming a user with voluminous output for a simple mistake. A command-line option may be provided
	to force full error output should it be desired.

	Additionally, an added benefit is ease of maintenance and localisation, by locating all error text
	in one place.

	To return an error, for example:

		return WithStackTrace(&ChecksumMismatchError{Actual: "123", Expected: "456"})

	To check to see if a specific error was returned, either of the following are acceptable:

		_, ok := err.(*ChecksumMismatchError)
		...

		switch err.(type) {
		case *ChecksumMismatchError:
        ...
*/

import (
	"bytes"
	"fmt"
	go_errors "github.com/go-errors/errors"
	"os"
	"runtime"
	"strings"
)

const (
	MaxStackDepth = 50
)

// Prefix applied to all error messages. Initialized in module init() function from os.Args[0]
var prefix string

/* CordCtlError is the interface for errors created by cordctl.
 *    ShouldDumpStack()
 *        Returns false for well-understood problems such as invalid user input where a brief error message is sufficient
 *		  Returns true for poorly-understood / unexpected problems where a full dump context may be useful
 *    Stack()
 *        Returns a string containing the stack trace where the error occurred
 *        Only useful if WithStackTrace() was called on the error
 */

type CordCtlError interface {
	error
	ShouldDumpStack() bool
	Stack() string
	AddStackTrace(skip int)
}

/* BaseError supports attaching stack traces to errors
 *    Borrowed the technique from github.com/go-errors. Decided against using go-errors directly since it requires
 *    wrapping our error classes. Instead, incorporated the stack trace directly into our error class.
 */

type BaseError struct {
	stack  []uintptr
	frames []go_errors.StackFrame
}

func (f *BaseError) AddStackTrace(skip int) {
	stack := make([]uintptr, MaxStackDepth)
	length := runtime.Callers(2+skip, stack[:])
	f.stack = stack[:length]
}

func (f *BaseError) Stack() string {
	buf := bytes.Buffer{}

	for _, frame := range f.StackFrames() {
		buf.WriteString(frame.String())
	}

	return string(buf.Bytes())
}

func (f *BaseError) StackFrames() []go_errors.StackFrame {
	if f.frames == nil {
		f.frames = make([]go_errors.StackFrame, len(f.stack))

		for i, pc := range f.stack {
			f.frames[i] = go_errors.NewStackFrame(pc)
		}
	}

	return f.frames
}

// ***************************************************************************
// UserError is composed into Errors that are due to user input

type UserError struct {
	BaseError
}

func (f UserError) ShouldDumpStack() bool {
	return false
}

// **************************************************************************
// TransferError is composed into Errors that are due to failures in transfers

type TransferError struct {
	BaseError
}

func (f TransferError) ShouldDumpStack() bool {
	return false
}

// ***************************************************************************
// UnexpectedError is things that we don't expect to happen. They should
// generate maximum error context, to provide useful information for developer
// diagnosis.

type UnexpectedError struct {
	BaseError
}

func (f UnexpectedError) ShouldDumpStack() bool {
	return true
}

// ***************************************************************************
// Specific error classes follow

// Checksum mismatch when downloading or uploading a file
type ChecksumMismatchError struct {
	TransferError
	Name     string // (optional) Name of file
	Expected string
	Actual   string
}

func (f ChecksumMismatchError) Error() string {
	if f.Name != "" {
		return fmt.Sprintf("%s %s: checksum mismatch (actual=%s, expected=%s)", prefix, f.Name, f.Expected, f.Actual)
	} else {
		return fmt.Sprintf("%s: checksum mismatch (actual=%s, expected=%s)", prefix, f.Expected, f.Actual)
	}
}

// User specified a model type that is not valid
type UnknownModelTypeError struct {
	UserError
	Name string // Name of model
}

func (f UnknownModelTypeError) Error() string {
	return fmt.Sprintf("%s: Model %s does not exist. Use `cordctl modeltype list` to get a list of available models", prefix, f.Name)
}

// User specified a model state that is not valid
type UnknownModelStateError struct {
	UserError
	Name string // Name of state
}

func (f UnknownModelStateError) Error() string {
	return fmt.Sprintf("%s: Model state %s does not exist", prefix, f.Name)
}

// Command requires a filter be specified
type FilterRequiredError struct {
	UserError
}

func (f FilterRequiredError) Error() string {
	return "Filter required. Use either an ID, --filter, or --all to specify which models to operate on"
}

// Command was aborted by the user
type AbortedError struct {
	UserError
}

func (f AbortedError) Error() string {
	return "Aborted"
}

// Command was aborted by the user
type NoMatchError struct {
	UserError
}

func (f NoMatchError) Error() string {
	return "No Match"
}

// User specified a field name that is not valid
type FieldDoesNotExistError struct {
	UserError
	ModelName string
	FieldName string
}

func (f FieldDoesNotExistError) Error() string {
	return fmt.Sprintf("%s: Model %s does not have field %s", prefix, f.ModelName, f.FieldName)
}

// User specified a query string that is not properly formatted
type IllegalQueryError struct {
	UserError
	Query string
}

func (f IllegalQueryError) Error() string {
	return fmt.Sprintf("%s: Illegal query string %s", prefix, f.Query)
}

// We failed to type convert something that we thought should have converted
type TypeConversionError struct {
	UnexpectedError
	Source      string
	Destination string
}

func (f TypeConversionError) Error() string {
	return fmt.Sprintf("%s: Failed to type convert from %s to %s", prefix, f.Source, f.Destination)
}

// Version did not match a constraint
type VersionConstraintError struct {
	UserError
	Name       string
	Version    string
	Constraint string
}

func (f VersionConstraintError) Error() string {
	return fmt.Sprintf("%s: %s version %s did not match constraint '%s'", prefix, f.Name, f.Version, f.Constraint)
}

// A model was not found
type ModelNotFoundError struct {
	UserError
	ModelName string
	Id        int32
	Queries   map[string]string
}

func (f ModelNotFoundError) Error() string {
	if f.Queries != nil {
		return fmt.Sprintf("%s: %s query %v not Found", prefix, f.ModelName, f.Queries)
	} else {
		return fmt.Sprintf("%s: Model %s id %d not Found", prefix, f.ModelName, f.Id)
	}
}

// InvalidInputError is a catch-all for user mistakes that aren't covered elsewhere
type InvalidInputError struct {
	UserError
	Message string
}

func (f InvalidInputError) Error() string {
	return fmt.Sprintf("%s: %s", prefix, f.Message)
}

func NewInvalidInputError(format string, params ...interface{}) *InvalidInputError {
	msg := fmt.Sprintf(format, params...)
	err := &InvalidInputError{Message: msg}
	err.AddStackTrace(2)
	return err
}

// InternalError is a catch-all for errors that don't fit somewhere else
type InternalError struct {
	UnexpectedError
	Message string
}

func (f InternalError) Error() string {
	return fmt.Sprintf("%s: %s", prefix, f.Message)
}

func NewInternalError(format string, params ...interface{}) *InternalError {
	msg := fmt.Sprintf(format, params...)
	err := &InternalError{Message: msg}
	err.AddStackTrace(2)
	return err
}

// ***************************************************************************
// Global exported function declarations

// Attach a stack trace to an error. The error passed in must be a pointer to an error structure for the
// CordCtlError interface to match.
func WithStackTrace(err CordCtlError) error {
	err.AddStackTrace(2)
	return err
}

// Set the prefix rather than using os.Args[0]. This is useful for testing.
func SetPrefix(s string) {
	prefix = s
}

// Convert an RPC error into a Cord Error
func RpcErrorWithIdToCordError(err error, modelName string, id int32) error {
	if err == nil {
		return err
	}
	if strings.Contains(err.Error(), "rpc error: code = NotFound") {
		err := &ModelNotFoundError{ModelName: modelName, Id: id}
		err.AddStackTrace(2)
		return err
	}
	return err
}

// Module initialization. Automatically defaults prefix to program name
func init() {
	prefix = os.Args[0]
}
