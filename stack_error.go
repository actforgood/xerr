// Copyright The ActForGood Authors.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://github.com/actforgood/xerr/blob/main/LICENSE.

package xerr

import (
	"fmt"
	"io"
	"runtime"
	"strconv"
)

// maxStackFrames is the maximum depth of callstack.
const maxStackFrames = 32

// stackError is an error enriched with callstack.
type stackError struct {
	// origErr is the original error, if this error wraps another one.
	origErr error
	// stackPCs holds the callstack program counters.
	stackPCs []uintptr
	// msg is this error's message.
	msg string
}

// Error returns the error's message.
// Implements std error interface.
//
// The returned value has the form <stackError.msg>: <stackError.origErr.Error()>,
// any of the 2 parts may be missing.
func (err stackError) Error() string {
	message := err.msg
	if err.origErr != nil {
		if message == "" {
			message = err.origErr.Error()
		} else {
			message += ": " + err.origErr.Error()
		}
	}

	return message
}

// Format implements [fmt.Formatter].
// The following verbs are supported:
//
//	%s    print the error. If the error has an original error, it will be
//	      printed recursively.
//	%v    same behaviour as %s.
//	%+v   extended format. Each frame of the error's call stack will
//	      be printed in detail.
func (err stackError) Format(f fmt.State, verb rune) {
	switch verb {
	case 'v':
		if f.Flag('+') {
			err.writeMsg(f)
			for _, pc := range err.stackPCs {
				fnName, file, line := getFrame(pc - 1)
				if !skipFrame(fnName, file) {
					writeFrame(f, fnName, file, line)
				}
			}

			return
		}

		fallthrough
	case 's':
		err.writeMsg(f)
	}
}

// writeMsg writes the error message.
// Used this instead of directly io.WriteString(w, err.Error()) to save some extra memory allocation.
func (err stackError) writeMsg(w io.Writer) {
	_, _ = io.WriteString(w, err.msg)
	if err.origErr != nil {
		if err.msg != "" {
			_, _ = io.WriteString(w, ": ")
		}
		_, _ = io.WriteString(w, err.origErr.Error())
	}
}

// Unwrap returns original error (can be nil).
// It implements [errors.Is] / [errors.As] APIs.
func (err stackError) Unwrap() error {
	return err.origErr
}

// New returns an error with the supplied message.
// New also records the stack trace at the point it was called.
func New(msg string) error {
	return &stackError{
		msg:      msg,
		stackPCs: getCallStack(maxStackFrames),
	}
}

// Errorf formats according to a format specifier and returns the string
// as a value that satisfies error.
// Errorf also records the stack trace at the point it was called.
func Errorf(format string, args ...interface{}) error {
	return &stackError{
		msg:      fmt.Sprintf(format, args...),
		stackPCs: getCallStack(maxStackFrames),
	}
}

// Wrap returns an error annotating err with a stack trace
// at the point Wrap is called, and the supplied message.
// If err is nil, Wrap returns nil.
// If err is another stack trace aware error, the final stack trace will
// consists of original error's stack trace + 1 trace of current Wrap call.
func Wrap(err error, msg string) error {
	if err == nil {
		return nil
	}

	var stackPCs []uintptr
	if sErr, ok := err.(*stackError); ok {
		stackPCs = append(getCallStack(1), sErr.stackPCs...)
	} else {
		stackPCs = getCallStack(maxStackFrames)
	}

	return &stackError{
		origErr:  err,
		msg:      msg,
		stackPCs: stackPCs,
	}
}

// Wrapf returns an error annotating err with a stack trace
// at the point Wrapf is called, and the message formatted according to a
// format specifier.
// If err is nil, Wrapf returns nil.
// If err is another stack trace aware error, the final stack trace will
// consists of original error's stack trace + 1 trace of current Wrapf call.
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	var stackPCs []uintptr
	if sErr, ok := err.(*stackError); ok {
		stackPCs = append(getCallStack(1), sErr.stackPCs...)
	} else {
		stackPCs = getCallStack(maxStackFrames)
	}

	return &stackError{
		origErr:  err,
		msg:      fmt.Sprintf(format, args...),
		stackPCs: stackPCs,
	}
}

// getCallStack return a slice of program counters of function invocations
// on the calling goroutine's stack.
func getCallStack(maxDepth int) []uintptr {
	pcs := make([]uintptr, maxDepth)
	n := runtime.Callers(3, pcs)

	return pcs[:n]
}

// writeFrame writes the given frame to the specified writer.
//
// The format in which is written is:
//
//	<functionName>
//	  <file>:<line>
//
// Example:
//
//	github.com/actforgood/xerr_test.TestX
//	  /Users/bogdan/work/go/xerr/errors_test.go:68
func writeFrame(w io.Writer, fnName string, file string, line int) {
	_, _ = io.WriteString(w, "\n")
	if frameFnNameProcessor != nil {
		_, _ = io.WriteString(w, frameFnNameProcessor(fnName))
	} else {
		_, _ = io.WriteString(w, fnName)
	}
	_, _ = io.WriteString(w, "\n\t")
	_, _ = io.WriteString(w, file)
	_, _ = io.WriteString(w, ":")
	_, _ = io.WriteString(w, strconv.FormatInt(int64(line), 10))
}

// getFrame returns function, file and line for a program counter.
func getFrame(pc uintptr) (fnName string, file string, line int) {
	fn := runtime.FuncForPC(pc)
	if fn != nil {
		fnName = fn.Name()
		file, line = fn.FileLine(pc)
	}

	return
}
