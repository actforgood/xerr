// Copyright 2022 Bogdan Constantinescu.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://github.com/actforgood/xerr/LICENSE.

package xerr_test

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"testing"

	"github.com/actforgood/xerr"
)

// Note: don't t.Parallel()ize tests as we share some global configuration.

func TestNew(t *testing.T) {
	// arrange
	var (
		subject = xerr.New
		regexes = []string{
			"something went bad\n",
			`github\.com/actforgood/xerr_test\.TestNew\n\t.+stack_error_test\.go:33`,
			`testing.tRunner\n\t.+testing.go:\d+`,
		}
	)

	// act
	resultErr := subject("something went bad")

	// assert
	if assertNotNil(t, resultErr) {
		assertEqual(t, "something went bad", resultErr.Error())
		assertEqual(t, "something went bad", fmt.Sprintf("%s", resultErr))
		assertEqual(t, "something went bad", fmt.Sprintf("%v", resultErr))
		errMsgWithStack := fmt.Sprintf("%+v", resultErr)
		for _, reg := range regexes {
			matched, _ := regexp.MatchString(reg, errMsgWithStack)
			if !assertTrue(t, matched) {
				t.Log("regex", reg, "errMsgWithStack", errMsgWithStack)
			}
		}
	}
}

func TestErrorf(t *testing.T) {
	// arrange
	var (
		subject = xerr.Errorf
		regexes = []string{
			"something went bad\n",
			`github\.com/actforgood/xerr_test\.TestErrorf\n\t.+stack_error_test\.go:62`,
			`testing.tRunner\n\t.+testing.go:\d+`,
		}
	)

	// act
	resultErr := subject("something %s %s", "went", "bad")

	// assert
	if assertNotNil(t, resultErr) {
		assertEqual(t, "something went bad", resultErr.Error())
		assertEqual(t, "something went bad", fmt.Sprintf("%s", resultErr))
		assertEqual(t, "something went bad", fmt.Sprintf("%v", resultErr))
		errMsgWithStack := fmt.Sprintf("%+v", resultErr)
		for _, reg := range regexes {
			matched, _ := regexp.MatchString(reg, errMsgWithStack)
			if !assertTrue(t, matched) {
				t.Log("regex", reg, "errMsgWithStack", errMsgWithStack)
			}
		}
	}
}

func TestWrap(t *testing.T) {
	t.Run("with standard error", testWrapWithStandardError)
	t.Run("with stack error", testWrapWithStackError)
	t.Run("with nil error", testWrapWithNilError)
	t.Run("with no message", testWrapWithNoMessage)
}

func testWrapWithStandardError(t *testing.T) {
	// arrange
	var (
		subject = xerr.Wrap
		origErr = errors.New("some standard error")
		regexes = []string{
			"something went bad: some standard error\n",
			`github\.com/actforgood/xerr_test\.testWrapWithStandardError\n\t.+stack_error_test\.go:99`,
			`testing.tRunner\n\t.+testing.go:\d+`,
		}
	)

	// act
	resultErr := subject(origErr, "something went bad")

	// assert
	if assertNotNil(t, resultErr) {
		assertEqual(t, "something went bad: some standard error", resultErr.Error())
		assertEqual(t, "something went bad: some standard error", fmt.Sprintf("%s", resultErr))
		assertEqual(t, "something went bad: some standard error", fmt.Sprintf("%v", resultErr))
		errMsgWithStack := fmt.Sprintf("%+v", resultErr)
		for _, reg := range regexes {
			matched, _ := regexp.MatchString(reg, errMsgWithStack)
			if !assertTrue(t, matched) {
				t.Log("regex", reg, "errMsgWithStack", errMsgWithStack)
			}
		}
	}
}

func testWrapWithStackError(t *testing.T) {
	// arrange
	var (
		subject = xerr.Wrap
		origErr = xerr.New("some error with stack")
		regexes = []string{
			"something went bad: some error with stack\n",
			`github\.com/actforgood/xerr_test\.testWrapWithStackError\n\t.+stack_error_test\.go:130`,
			`github\.com/actforgood/xerr_test\.testWrapWithStackError\n\t.+stack_error_test\.go:120`,
			`testing.tRunner\n\t.+testing.go:\d+`,
		}
	)

	// act
	resultErr := subject(origErr, "something went bad")

	// assert
	if assertNotNil(t, resultErr) {
		assertEqual(t, "something went bad: some error with stack", resultErr.Error())
		assertEqual(t, "something went bad: some error with stack", fmt.Sprintf("%s", resultErr))
		assertEqual(t, "something went bad: some error with stack", fmt.Sprintf("%v", resultErr))
		errMsgWithStack := fmt.Sprintf("%+v", resultErr)
		for _, reg := range regexes {
			matched, _ := regexp.MatchString(reg, errMsgWithStack)
			if !assertTrue(t, matched) {
				t.Log("debug", "regex", reg, "errMsgWithStack", errMsgWithStack)
			}
		}
	}
}

func testWrapWithNilError(t *testing.T) {
	// arrange
	var (
		subject = xerr.Wrap
		origErr error
	)

	// act
	resultErr := subject(origErr, "something went bad")

	// assert
	assertNil(t, resultErr)
}

func testWrapWithNoMessage(t *testing.T) {
	// arrange
	var (
		subject = xerr.Wrap
		origErr = errors.New("some standard error")
		regexes = []string{
			"some standard error\n",
			`github\.com/actforgood/xerr_test\.testWrapWithNoMessage\n\t.+stack_error_test\.go:174`,
			`testing.tRunner\n\t.+testing.go:\d+`,
		}
	)

	// act
	resultErr := subject(origErr, "")

	// assert
	if assertNotNil(t, resultErr) {
		assertEqual(t, "some standard error", resultErr.Error())
		assertEqual(t, "some standard error", fmt.Sprintf("%s", resultErr))
		assertEqual(t, "some standard error", fmt.Sprintf("%v", resultErr))
		errMsgWithStack := fmt.Sprintf("%+v", resultErr)
		for _, reg := range regexes {
			matched, _ := regexp.MatchString(reg, errMsgWithStack)
			if !assertTrue(t, matched) {
				t.Log("regex", reg, "errMsgWithStack", errMsgWithStack)
			}
		}
	}
}

func TestWrapf(t *testing.T) {
	t.Run("with standard error", testWrapfWithStandardError)
	t.Run("with stack error", testWrapfWithStackError)
	t.Run("with nil error", testWrapfWithNilError)
}

func testWrapfWithStandardError(t *testing.T) {
	// arrange
	var (
		subject = xerr.Wrapf
		origErr = errors.New("some standard error")
		regexes = []string{
			"something went bad: some standard error\n",
			`github\.com/actforgood/xerr_test\.testWrapfWithStandardError\n\t.+stack_error_test\.go:210`,
			`testing.tRunner\n\t.+testing.go:\d+`,
		}
	)

	// act
	resultErr := subject(origErr, "something %s %s", "went", "bad")

	// assert
	if assertNotNil(t, resultErr) {
		assertEqual(t, "something went bad: some standard error", resultErr.Error())
		assertEqual(t, "something went bad: some standard error", fmt.Sprintf("%s", resultErr))
		assertEqual(t, "something went bad: some standard error", fmt.Sprintf("%v", resultErr))
		errMsgWithStack := fmt.Sprintf("%+v", resultErr)
		for _, reg := range regexes {
			matched, _ := regexp.MatchString(reg, errMsgWithStack)
			if !assertTrue(t, matched) {
				t.Log("debug", "regex", reg, "errMsgWithStack", errMsgWithStack)
			}
		}
	}
}

func testWrapfWithStackError(t *testing.T) {
	// arrange
	var (
		subject = xerr.Wrapf
		origErr = xerr.New("some error with stack")
		regexes = []string{
			"something went bad: some error with stack\n",
			`github\.com/actforgood/xerr_test\.testWrapfWithStackError\n\t.+stack_error_test\.go:241`,
			`github\.com/actforgood/xerr_test\.testWrapfWithStackError\n\t.+stack_error_test\.go:231`,
			`testing.tRunner\n\t.+testing.go:\d+`,
		}
	)

	// act
	resultErr := subject(origErr, "something %s %s", "went", "bad")

	// assert
	if assertNotNil(t, resultErr) {
		assertEqual(t, "something went bad: some error with stack", resultErr.Error())
		assertEqual(t, "something went bad: some error with stack", fmt.Sprintf("%s", resultErr))
		assertEqual(t, "something went bad: some error with stack", fmt.Sprintf("%v", resultErr))
		errMsgWithStack := fmt.Sprintf("%+v", resultErr)
		for _, reg := range regexes {
			matched, _ := regexp.MatchString(reg, errMsgWithStack)
			if !assertTrue(t, matched) {
				t.Log("debug", "regex", reg, "errMsgWithStack", errMsgWithStack)
			}
		}
	}
}

func testWrapfWithNilError(t *testing.T) {
	// arrange
	var (
		subject = xerr.Wrapf
		origErr error
	)

	// act
	resultErr := subject(origErr, "something went bad")

	// assert
	assertNil(t, resultErr)
}

func TestUnwrap(t *testing.T) {
	// arrange
	var (
		stackErr = xerr.New("some error with stack trace")
		stdErr   = errors.New("some standard error")
		tests    = [...]struct {
			name      string
			subject   error
			targetErr error
			expected  bool
		}{
			{
				name:      "wrap(stdErr), is stdErr",
				subject:   xerr.Wrap(stdErr, "wrap"),
				targetErr: stdErr,
				expected:  true,
			},
			{
				name:      "wrap(stdErr), is not another stdErr",
				subject:   xerr.Wrap(stdErr, "wrap"),
				targetErr: io.EOF,
				expected:  false,
			},
			{
				name:      "wrap(wrap(stdErr)), is stdErr",
				subject:   xerr.Wrap(xerr.Wrap(stdErr, "1st wrap"), "2nd wrap"),
				targetErr: stdErr,
				expected:  true,
			},
			{
				name:      "wrap(stackErr), is stackErr",
				subject:   xerr.Wrap(stackErr, "wrap"),
				targetErr: stackErr,
				expected:  true,
			},
			{
				name:      "wrap(stackErr), is not stdErr",
				subject:   xerr.Wrap(stackErr, "wrap"),
				targetErr: stdErr,
				expected:  false,
			},
			{
				name:      "wrap(wrap(wrap(stackErr))), is stackErr",
				subject:   xerr.Wrap(xerr.Wrap(xerr.Wrap(stackErr, "1st wrap"), "2nd wrap"), "3rd wrap"),
				targetErr: stackErr,
				expected:  true,
			},
			{
				name:      "stackErr is stackErr",
				subject:   stackErr,
				targetErr: stackErr,
				expected:  true,
			},
			{
				name:      "stackErr is not srdErr",
				subject:   stackErr,
				targetErr: stdErr,
				expected:  false,
			},
			{
				name:      "nil is not stackErr",
				subject:   nil,
				targetErr: stackErr,
				expected:  false,
			},
			{
				name:      "stackErr is not nil",
				subject:   stackErr,
				targetErr: nil,
				expected:  false,
			},
		}
	)

	for _, testData := range tests {
		test := testData // capture range variable
		t.Run(test.name, func(t *testing.T) {
			// act
			result := errors.Is(test.subject, test.targetErr)

			// assert
			assertEqual(t, test.expected, result)
		})
	}
}

func TestNew_withGlobalConfigurationChanged(t *testing.T) {
	// arrange
	var (
		subject = xerr.New
		regexes = []string{
			"something went bad\n",
			`github\.com/actforgood/xerr_test\.TestNew_withGlobalConfigurationChanged_FOO\n\t.+stack_error_test\.go:\d+`,
		}
		skipFrameCallsCnt            = 0
		frameFnNameProcessorCallsCnt = 0
	)
	xerr.SetSkipFrame(func(fnName, file string) bool {
		skipFrameCallsCnt++

		return !strings.HasPrefix(fnName, "github.com/actforgood")
	})
	xerr.SetFrameFnNameProcessor(func(fnName string) string {
		frameFnNameProcessorCallsCnt++

		return fnName + "_FOO"
	})
	defer func() { // restore original global state
		xerr.SetSkipFrame(xerr.AllowFrame)
		xerr.SetFrameFnNameProcessor(nil)
	}()

	// act
	resultErr := subject("something went bad")

	// assert
	if assertNotNil(t, resultErr) {
		assertEqual(t, "something went bad", resultErr.Error())
		assertEqual(t, "something went bad", fmt.Sprintf("%s", resultErr))
		assertEqual(t, "something went bad", fmt.Sprintf("%v", resultErr))
		errMsgWithStack := fmt.Sprintf("%+v", resultErr)
		for _, reg := range regexes {
			matched, _ := regexp.MatchString(reg, errMsgWithStack)
			if !assertTrue(t, matched) {
				t.Log("regex", reg, "errMsgWithStack", errMsgWithStack)
			}
		}
		matched, _ := regexp.MatchString(`testing.tRunner\n\t.+testing.go:\d+`, errMsgWithStack)
		if !assertFalse(t, matched) {
			t.Log("regex", `testing.tRunner\n\t.+testing.go:\d+`, "errMsgWithStack", errMsgWithStack)
		}

		assertTrue(t, skipFrameCallsCnt >= 1)
		assertTrue(t, frameFnNameProcessorCallsCnt >= 1)
	}
}

func BenchmarkNew(b *testing.B) {
	for n := 0; n < b.N; n++ {
		err := xerr.New("some error with stack trace")
		_ = fmt.Sprintf("%+v", err)
	}
}

func BenchmarkWrap_withStandardError(b *testing.B) {
	origErr := errors.New("some standard error")

	for n := 0; n < b.N; n++ {
		err := xerr.Wrap(origErr, "wrap")
		_ = fmt.Sprintf("%+v", err)
	}
}

func BenchmarkWrap_withStackError(b *testing.B) {
	origErr := xerr.New("some error with stack trace")

	for n := 0; n < b.N; n++ {
		err := xerr.Wrap(origErr, "wrap")
		_ = fmt.Sprintf("%+v", err)
	}
}
