// Copyright The ActForGood Authors.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://github.com/actforgood/xerr/blob/main/LICENSE.

package xerr_test

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"testing"
	"testing/synctest"

	"github.com/actforgood/xerr"
)

func TestMultiError_initializedFromStart(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		subject    = xerr.NewMultiError()
		customErr  = dummyCustomError{}
		extractErr dummyCustomError
		stdErr1    = errors.New("some standard error 1")
		stdErr2    = errors.New("some standard error 2")
		stdErr3    = errors.New("some standard error 3")
	)

	// act & assert
	// test initial state
	assertNil(t, subject.ErrOrNil())
	assertNotNil(t, subject.Errors())
	assertEqual(t, 0, len(subject.Errors()))
	assertEqual(t, "", subject.Error())
	assertEqual(t, "", fmt.Sprintf("%+v", subject))
	assertFalse(t, errors.Is(subject, stdErr3))
	assertFalse(t, errors.As(subject, &extractErr))

	// add an error
	subject = subject.Add(stdErr1)
	assertNotNil(t, subject.ErrOrNil())
	assertEqual(t, stdErr1.Error(), subject.Error())
	errs := subject.Errors()
	assertEqual(t, []error{stdErr1}, errs)

	// add nil
	subject = subject.Add(nil)
	assertNotNil(t, subject.ErrOrNil())
	assertEqual(t, stdErr1.Error(), subject.Error())
	assertEqual(t, []error{stdErr1}, subject.Errors())

	// add another error
	subject = subject.Add(stdErr2)
	assertNotNil(t, subject.ErrOrNil())
	assertEqual(t, stdErr1.Error()+"\n"+stdErr2.Error(), subject.Error())
	assertEqual(t, 1, len(errs)) // see we got a copy the first time
	assertEqual(t, []error{stdErr1, stdErr2}, subject.Errors())

	// add unique an already existing error
	subject = subject.AddOnce(stdErr1)
	assertNotNil(t, subject.ErrOrNil())
	assertEqual(t, stdErr1.Error()+"\n"+stdErr2.Error(), subject.Error())
	assertEqual(t, []error{stdErr1, stdErr2}, subject.Errors())

	// add unique a new error
	subject = subject.AddOnce(customErr)
	assertNotNil(t, subject.ErrOrNil())
	assertEqual(t, stdErr1.Error()+"\n"+stdErr2.Error()+"\n"+customErr.Error(), subject.Error())
	assertEqual(t, []error{stdErr1, stdErr2, customErr}, subject.Errors())

	// add unique a nil error and a new one
	subject = subject.AddOnce(customErr, nil, stdErr3)
	assertNotNil(t, subject.ErrOrNil())
	assertEqual(
		t,
		stdErr1.Error()+"\n"+stdErr2.Error()+"\n"+customErr.Error()+"\n"+stdErr3.Error(),
		subject.Error(),
	)
	assertEqual(t, []error{stdErr1, stdErr2, customErr, stdErr3}, subject.Errors())

	// test Is/As/Unwrap
	assertTrue(t, errors.Is(subject, subject))
	assertTrue(t, errors.Is(subject, stdErr1))
	assertTrue(t, errors.Is(subject, stdErr2))
	assertTrue(t, errors.Is(subject, stdErr3))
	assertFalse(t, errors.Is(subject, nil))
	assertFalse(t, errors.Is(subject, errors.New("a different error")))

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

func TestMultiError_initializedLater(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		subject    *xerr.MultiError
		customErr  = dummyCustomError{}
		extractErr dummyCustomError
		stdErr1    = errors.New("some standard error 1")
		stdErr2    = errors.New("some standard error 2")
		stdErr3    = errors.New("some standard error 3")
	)

	// act & assert
	// test initial state
	assertNil(t, subject.ErrOrNil())
	assertNil(t, subject.Errors())
	assertEqual(t, "", subject.Error())
	assertEqual(t, "", fmt.Sprintf("%+v", subject))
	assertFalse(t, errors.Is(subject, stdErr3))
	assertFalse(t, errors.As(subject, &extractErr))

	// add an error
	subject = subject.Add(stdErr1)
	assertNotNil(t, subject.ErrOrNil())
	assertEqual(t, stdErr1.Error(), subject.Error())
	errs := subject.Errors()
	assertEqual(t, []error{stdErr1}, errs)

	// add nil
	subject = subject.Add(nil)
	assertNotNil(t, subject.ErrOrNil())
	assertEqual(t, stdErr1.Error(), subject.Error())
	assertEqual(t, []error{stdErr1}, subject.Errors())

	// add another error
	subject = subject.Add(stdErr2)
	assertNotNil(t, subject.ErrOrNil())
	assertEqual(t, stdErr1.Error()+"\n"+stdErr2.Error(), subject.Error())
	assertEqual(t, 1, len(errs)) // see we got a copy the first time
	assertEqual(t, []error{stdErr1, stdErr2}, subject.Errors())

	// add unique an already existing error
	subject = subject.AddOnce(stdErr1)
	assertNotNil(t, subject.ErrOrNil())
	assertEqual(t, stdErr1.Error()+"\n"+stdErr2.Error(), subject.Error())
	assertEqual(t, []error{stdErr1, stdErr2}, subject.Errors())

	// add unique a new error
	subject = subject.AddOnce(customErr)
	assertNotNil(t, subject.ErrOrNil())
	assertEqual(t, stdErr1.Error()+"\n"+stdErr2.Error()+"\n"+customErr.Error(), subject.Error())
	assertEqual(t, []error{stdErr1, stdErr2, customErr}, subject.Errors())

	// add unique a nil error and a new one
	subject = subject.AddOnce(customErr, nil, stdErr3)
	assertNotNil(t, subject.ErrOrNil())
	assertEqual(
		t,
		stdErr1.Error()+"\n"+stdErr2.Error()+"\n"+customErr.Error()+"\n"+stdErr3.Error(),
		subject.Error(),
	)
	assertEqual(t, []error{stdErr1, stdErr2, customErr, stdErr3}, subject.Errors())

	// test Is/As/Unwrap
	assertTrue(t, errors.Is(subject, subject))
	assertTrue(t, errors.Is(subject, stdErr1))
	assertTrue(t, errors.Is(subject, stdErr2))
	assertTrue(t, errors.Is(subject, stdErr3))
	assertFalse(t, errors.Is(subject, nil))
	assertFalse(t, errors.Is(subject, errors.New("a different error")))

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

type dummyCustomError struct{}

func (dummyCustomError) Error() string { return "dummy custom error" }
func (dumErr dummyCustomError) Format(f fmt.State, verb rune) {
	if verb == 'v' && f.Flag('+') {
		_, _ = io.WriteString(f, dumErr.Error()+" %+v formatted")
	}
}

func TestMultiError_Reset(t *testing.T) {
	t.Parallel()

	// act & assert - subject not initialized
	var subject *xerr.MultiError
	subject.Reset()
	assertNil(t, subject.Errors())

	// act & assert - subject is initialized, has no errors
	subject = xerr.NewMultiError()
	subject.Reset()
	assertNotNil(t, subject.Errors())
	assertEqual(t, 0, len(subject.Errors()))

	// act & assert - subject with errors
	_ = subject.AddOnce(io.ErrUnexpectedEOF)
	assertEqual(t, 1, len(subject.Errors()))
	subject.Reset()
	assertNotNil(t, subject.Errors())
	assertEqual(t, 0, len(subject.Errors()))
}

func TestMultiError_Unwrap_Is(t *testing.T) {
	t.Parallel()

	// arrange & act & assert
	subject := xerr.NewMultiError()
	assertTrue(t, errors.Is(subject, subject))

	// arrange & act & assert
	_ = subject.Add(io.ErrUnexpectedEOF)
	assertTrue(t, errors.Is(subject, subject))
	assertTrue(t, errors.Is(subject, io.ErrUnexpectedEOF))

	// arrange & act & assert
	_ = subject.Add(io.ErrShortWrite)
	assertTrue(t, errors.Is(subject, subject))
	assertTrue(t, errors.Is(subject, io.ErrUnexpectedEOF))
	assertTrue(t, errors.Is(subject, io.ErrShortWrite))
}

func TestMultiError_concurrency(t *testing.T) {
	t.Parallel()

	synctest.Test(t, func(*testing.T) {
		// arrange
		var (
			subject          = xerr.NewMultiError()
			goroutinesNo     = 200
			extractThreadReg = regexp.MustCompile(`\d+`)
		)

		// act
		for i := range goroutinesNo {
			go func(mErr *xerr.MultiError, threadNo int) {
				err := errors.New("err from threadNo " + strconv.FormatInt(int64(threadNo+1), 10))
				// perform all kind of ops upon subject that can trigger race conditions when running t with -race
				_ = mErr.Add(err)
				_ = mErr.AddOnce(err)
				_ = mErr.Errors()
				_ = mErr.Error()
				assertNotNil(t, mErr.ErrOrNil())
				_ = fmt.Sprintf("%+v", mErr)
				assertTrue(t, errors.Is(mErr, err))
			}(subject, i)
		}
		synctest.Wait()

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
	})
}

func BenchmarkMultiError_concurrentSafe(b *testing.B) {
	var (
		err  = errors.New("some error to be Added to MultiError")
		mErr = xerr.NewMultiError()
	)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = mErr.Add(err)
			_ = mErr.AddOnce(err)
			_ = mErr.Errors()
			_ = mErr.Error()
			_ = mErr.ErrOrNil()
			_ = errors.Is(mErr, err)
		}
	})
}

func BenchmarkMultiError_notConcurrentSafe(b *testing.B) {
	var (
		err  = errors.New("some error to be Added to MultiError")
		mErr *xerr.MultiError
	)

	for range b.N {
		mErr = mErr.Add(err)
		mErr = mErr.AddOnce(err)
		_ = mErr.Errors()
		_ = mErr.Error()
		_ = mErr.ErrOrNil()
		_ = errors.Is(mErr, err)
	}
}

func BenchmarkMultiError_Reset(b *testing.B) {
	var (
		err  = errors.New("some error to be Added to MultiError")
		mErr = xerr.NewMultiError()
	)

	for range b.N {
		_ = mErr.Add(err)
		mErr.Reset()
	}
}
