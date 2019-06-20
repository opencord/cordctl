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

	    return ChecksumMismatchError{Actual: "123", Expected: "456"}

	To check to see if a specific error was returned, either of the following are acceptable:

		_, ok := err.(ChecksumMismatchError)
		...

		switch err.(type) {
		case ChecksumMismatchError:
		...
*/

import (
	"fmt"
)

const (
	PREFIX = "cordctl" // A prefix that appears at the start of all error messages
)

/* CordCtlError is the interface for errors created by cordctl.
 *    ShouldDumpStack()
 *        returns false for well-understood problems such as invalid user input where a brief error message is sufficient
 *		  returns true for poorly-understood / unexpected problems where a full dump context may be useful
 */

type CordCtlError interface {
	error
	ShouldDumpStack() bool
}

// ***************************************************************************
// UserError is composed into Errors that are due to user input

type UserError struct {
}

func (f UserError) ShouldDumpStack() bool {
	return false
}

// **************************************************************************
// TransferError is composed into Errors that are due to failures in transfers

type TransferError struct {
}

func (f TransferError) ShouldDumpStack() bool {
	return false
}

// ***************************************************************************
// UnexpectedError is things that we don't expect to happen. They should
// generate maximum error context, to provide useful information for developer
// diagnosis.

type UnexpectedError struct {
}

func (f UnexpectedError) ShouldDumpStack() bool {
	return true
}

// ***************************************************************************
// Specific error classes follow

type ChecksumMismatchError struct {
	UserError
	Name     string // (optional) Name of file
	Expected string
	Actual   string
}

func (f ChecksumMismatchError) Error() string {
	if f.Name != "" {
		return fmt.Sprintf("%s %s: checksum mismatch (actual=%s, expected=%s)", PREFIX, f.Name, f.Expected, f.Actual)
	} else {
		return fmt.Sprintf("%s: checksum mismatch (actual=%s, expected=%s)", PREFIX, f.Expected, f.Actual)
	}
}

type UnknownModelTypeError struct {
	UserError
	Name string // Name of model
}

func (f UnknownModelTypeError) Error() string {
	return fmt.Sprintf("%s: Model %s does not exist. Use `cordctl modeltype list` to get a list of available models", PREFIX, f.Name)
}
