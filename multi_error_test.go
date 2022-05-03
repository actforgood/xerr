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
	"strconv"
	"sync"
	"testing"

	"github.com/actforgood/xerr"
)

func TestMultiError(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		subject   = xerr.NewMultiError()
		customErr = dummyCustomErr{}
		stdErr1   = errors.New("some standard error 1")
		stdErr2   = errors.New("some standard error 2")
		stdErr3   = errors.New("some standard error 3")
	)

	// act & assert
	// test initial state
	assertNil(t, subject.ErrOrNil())
	assertEqual(t, "", subject.Error())
	assertEqual(t, "", fmt.Sprintf("%+v", subject))

	// add an error
	subject.Add(stdErr1)
	assertNotNil(t, subject.ErrOrNil())
	assertEqual(t, stdErr1.Error(), subject.Error())
	errs := subject.Errors()
	assertEqual(t, []error{stdErr1}, errs)

	// add nil
	subject.Add(nil)
	assertNotNil(t, subject.ErrOrNil())
	assertEqual(t, stdErr1.Error(), subject.Error())
	assertEqual(t, []error{stdErr1}, subject.Errors())

	// add another error
	subject.Add(stdErr2)
	assertNotNil(t, subject.ErrOrNil())
	assertEqual(t, stdErr1.Error()+"\n"+stdErr2.Error(), subject.Error())
	assertEqual(t, 1, len(errs)) // see we got a copy the first time
	assertEqual(t, []error{stdErr1, stdErr2}, subject.Errors())

	// add unique an already existing error
	subject.AddOnce(stdErr1)
	assertNotNil(t, subject.ErrOrNil())
	assertEqual(t, stdErr1.Error()+"\n"+stdErr2.Error(), subject.Error())
	assertEqual(t, []error{stdErr1, stdErr2}, subject.Errors())

	// add unique a new error
	subject.AddOnce(customErr)
	assertNotNil(t, subject.ErrOrNil())
	assertEqual(t, stdErr1.Error()+"\n"+stdErr2.Error()+"\n"+customErr.Error(), subject.Error())
	assertEqual(t, []error{stdErr1, stdErr2, customErr}, subject.Errors())

	// add unique a nil error and a new one
	subject.AddOnce(customErr, nil, stdErr3)
	assertNotNil(t, subject.ErrOrNil())
	assertEqual(t, stdErr1.Error()+"\n"+stdErr2.Error()+"\n"+customErr.Error()+"\n"+stdErr3.Error(), subject.Error())
	assertEqual(t, []error{stdErr1, stdErr2, customErr, stdErr3}, subject.Errors())

	// test Is/As/Unwrap
	assertTrue(t, errors.Is(subject, subject))
	assertTrue(t, errors.Is(subject, stdErr1))
	assertTrue(t, errors.Is(subject, stdErr2))
	assertTrue(t, errors.Is(subject, stdErr3))
	assertFalse(t, errors.Is(subject, nil))
	assertFalse(t, errors.Is(subject, errors.New("a different error")))

	var extractErr dummyCustomErr
	assertTrue(t, errors.As(subject, &extractErr))
	assertEqual(t, customErr, extractErr)

	// test Format
	expectedFmtOutcome := `error #1
some standard error 1
error #2
some standard error 2
error #3
dummy custom error %+v formatted
error #4
some standard error 3`
	assertEqual(t, expectedFmtOutcome, fmt.Sprintf("%+v", subject))
}

type dummyCustomErr struct{}

func (dummyCustomErr) Error() string { return "dummy custom error" }
func (dumErr dummyCustomErr) Format(f fmt.State, verb rune) {
	if verb == 'v' && f.Flag('+') {
		_, _ = io.WriteString(f, dumErr.Error()+" %+v formatted")
	}
}

func TestMultiError_concurrency(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		subject          = xerr.NewMultiError()
		goroutinesNo     = 200
		wg               sync.WaitGroup
		extractThreadReg = regexp.MustCompile(`\d+`)
	)

	// act
	for i := 0; i < goroutinesNo; i++ {
		wg.Add(1)
		go func(mErr *xerr.MultiError, threadNo int) {
			defer wg.Done()

			err := errors.New("err from threadNo " + strconv.FormatInt(int64(threadNo+1), 10))
			// perform all kind of ops upon subject that can trigger race conditions when running t with -race
			mErr.Add(err)
			mErr.AddOnce(err)
			_ = mErr.Errors()
			_ = mErr.Error()
			assertNotNil(t, mErr.ErrOrNil())
			_ = fmt.Sprintf("%+v", mErr)
			assertTrue(t, errors.Is(mErr, err))
		}(subject, i)
	}
	wg.Wait()

	// assert
	assertNotNil(t, subject.ErrOrNil())
	assertEqual(t, goroutinesNo, len(subject.Errors()))
	sum := 0
	for _, err := range subject.Errors() {
		matches := extractThreadReg.FindAllString(err.Error(), 1)
		if len(matches) == 1 {
			if threadNo, err := strconv.Atoi(matches[0]); err == nil {
				sum += threadNo
			}
		}
	}
	assertEqual(t, goroutinesNo*(goroutinesNo+1)/2, sum)
}
