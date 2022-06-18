// Copyright 2022 Bogdan Constantinescu.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://github.com/actforgood/xerr/LICENSE.

package xerr

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"sync"
)

// MultiError holds a pool of errors.
// Its APIs are concurrent safe if you initialize it
// with NewMultiError().
type MultiError struct {
	errors []error
	mu     *sync.RWMutex
}

// NewMultiError instantiates a new MultiError object.
// Use it to initialize from start your MultiError variable
// if you use it in a concurrent context, or you need to pass it
// as parameter to a function. Otherwise just declare the variable's type
// and get effective instance returned by Add() / AddOnce() APIs,
// to avoid unnecessary allocation if those APIs end up never being called.
func NewMultiError() *MultiError {
	return &MultiError{
		errors: make([]error, 0),
		mu:     new(sync.RWMutex),
	}
}

// newMultiError initializes internally a MultiError object, not concurrent safe.
func newMultiError() *MultiError {
	return &MultiError{
		errors: make([]error, 0),
	}
}

// Error returns the error's message.
// Implements std error interface.
// Returns all stored errors' messages, new line separated.
func (mErr *MultiError) Error() string {
	if mErr == nil {
		return ""
	}
	mErr.rLock()
	defer mErr.rUnlock()

	switch len(mErr.errors) {
	case 0:
		return ""
	case 1:
		return mErr.errors[0].Error()
	default:
		buf := bytes.Buffer{}
		for _, err := range mErr.errors {
			buf.WriteString(err.Error())
			buf.WriteByte('\n')
		}

		return string(buf.Bytes()[:buf.Len()-1])
	}
}

// Add appends the given error(s) in MultiError.
// It returns the MultiError, eventually initialized.
func (mErr *MultiError) Add(errs ...error) *MultiError {
	for _, err := range errs {
		if err != nil {
			if mErr == nil {
				mErr = newMultiError()
			}
			mErr.lock()
			mErr.errors = append(mErr.errors, err)
			mErr.unlock()
		}
	}

	return mErr
}

// AddOnce stores the given error(s) in MultiError,
// only if they do not exist already. Comparison is
// accomplished with Is() API.
// It returns the MultiError, eventually initialized.
func (mErr *MultiError) AddOnce(errs ...error) *MultiError {
	for _, err := range errs {
		if err == nil {
			continue
		}
		if mErr == nil {
			mErr = newMultiError()
		}

		mErr.lock()
		if mErr.hasError(err) {
			mErr.unlock()

			continue
		}
		mErr.errors = append(mErr.errors, err)
		mErr.unlock()
	}

	return mErr
}

// hasError checks if an error already exists in MultiError.
// Comparison is done with Is() API.
func (mErr *MultiError) hasError(err error) bool {
	for _, storedErr := range mErr.errors {
		if errors.Is(storedErr, err) {
			return true
		}
	}

	return false
}

// Errors returns a copy of stored errors.
func (mErr *MultiError) Errors() []error {
	if mErr == nil {
		return nil
	}
	mErr.rLock()
	errors := make([]error, len(mErr.errors))
	copy(errors, mErr.errors)
	mErr.rUnlock()

	return errors
}

// Reset cleans up stored errors, if any.
func (mErr *MultiError) Reset() {
	if mErr == nil {
		return
	}

	mErr.lock()
	if len(mErr.errors) > 0 {
		// keep the allocated memory
		for idx := range mErr.errors {
			mErr.errors[idx] = nil
		}
		mErr.errors = mErr.errors[:0]
	}
	mErr.unlock()
}

// ErrOrNil returns nil if MultiError does not have any stored errors,
// or the single error it stores,
// or self if has more more than 1 error.
func (mErr *MultiError) ErrOrNil() error {
	if mErr == nil {
		return nil
	}
	mErr.rLock()
	defer mErr.rUnlock()

	switch len(mErr.errors) {
	case 0:
		return nil
	case 1:
		return mErr.errors[0]
	default:
		return mErr
	}
}

// Format implements fmt.Formatter.
// It relies upon individual error's Format() API if applicable,
// otherwise Error() 's outcome is taken into account.
func (mErr *MultiError) Format(f fmt.State, verb rune) {
	if mErr == nil {
		return
	}
	mErr.rLock()
	defer mErr.rUnlock()

	errorsLen := len(mErr.errors)
	if errorsLen == 0 {
		return
	}

	for idx, err := range mErr.errors {
		if verb == 'v' {
			_, _ = io.WriteString(f, "error #")
			_, _ = io.WriteString(f, strconv.FormatInt(int64(idx+1), 10))
			_, _ = io.WriteString(f, "\n")
		}
		if errFmt, ok := err.(fmt.Formatter); ok {
			errFmt.Format(f, verb)
		} else {
			_, _ = io.WriteString(f, err.Error())
		}
		if idx != errorsLen-1 {
			_, _ = io.WriteString(f, "\n")
		}
	}
}

// Unwrap returns original error (can be nil).
// It implements standard error Is()/As() APIs.
// Returns recursively first error from stored errors.
func (mErr *MultiError) Unwrap() error {
	if mErr == nil {
		return nil
	}
	mErr.rLock()
	defer mErr.rUnlock()

	if len(mErr.errors) == 0 {
		return nil
	} else if len(mErr.errors) == 1 {
		return mErr.errors[0]
	}

	var newMultiErr *MultiError
	if mErr.mu != nil {
		newMultiErr = NewMultiError()
	} else {
		newMultiErr = newMultiError()
	}
	_ = newMultiErr.Add(mErr.errors[1:]...)

	return newMultiErr
}

// As implements standard error As() API,
// comparing the first error from stored ones.
func (mErr *MultiError) As(target interface{}) bool {
	if mErr == nil {
		return false
	}
	mErr.rLock()
	defer mErr.rUnlock()

	if len(mErr.errors) > 0 {
		return errors.As(mErr.errors[0], target)
	}

	return false
}

// Is implements standard error Is() API,
// comparing the first error from stored ones.
func (mErr *MultiError) Is(target error) bool {
	if mErr == nil {
		return mErr == target
	}
	mErr.rLock()
	defer mErr.rUnlock()

	if len(mErr.errors) > 0 {
		return errors.Is(mErr.errors[0], target)
	}

	return false
}

func (mErr *MultiError) lock() {
	if mErr.mu != nil {
		mErr.mu.Lock()
	}
}

func (mErr *MultiError) unlock() {
	if mErr.mu != nil {
		mErr.mu.Unlock()
	}
}

func (mErr *MultiError) rLock() {
	if mErr.mu != nil {
		mErr.mu.RLock()
	}
}

func (mErr *MultiError) rUnlock() {
	if mErr.mu != nil {
		mErr.mu.RUnlock()
	}
}
