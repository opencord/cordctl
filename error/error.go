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
	return true
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

type UnknownModelTypeError struct {
	UserError
	Name string // Name of model
}

func (f UnknownModelTypeError) Error() string {
	return fmt.Sprintf("%s: Model %s does not exist. Use `cordctl modeltype list` to get a list of available models", prefix, f.Name)
}

func WithStackTrace(err CordCtlError) error {
	err.AddStackTrace(2)
	return err
}

func SetPrefix(s string) {
	prefix = s
}

func init() {
	prefix = os.Args[0]
}
