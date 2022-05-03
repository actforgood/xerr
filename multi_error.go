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
// Its APIs are concurrent safe.
type MultiError struct {
	errors []error
	mu     *sync.RWMutex
}

// NewMultiError instantiates a new MultiError object.
func NewMultiError() *MultiError {
	return &MultiError{
		errors: make([]error, 0),
		mu:     new(sync.RWMutex),
	}
}

// Error returns the error's message.
// Implements std error interface.
// Returns all stored errors' messages, new line separated.
func (mErr *MultiError) Error() string {
	mErr.mu.RLock()
	defer mErr.mu.RUnlock()

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
func (mErr *MultiError) Add(errs ...error) {
	for _, err := range errs {
		if err != nil {
			mErr.mu.Lock()
			mErr.errors = append(mErr.errors, err)
			mErr.mu.Unlock()
		}
	}
}

// AddOnce stores the given error(s) in MultiError,
// only if they do not exist already. Comparison is
// accomplished with Is() API.
func (mErr *MultiError) AddOnce(errs ...error) {
	for _, err := range errs {
		if err == nil {
			continue
		}

		mErr.mu.Lock()
		if mErr.hasError(err) {
			mErr.mu.Unlock()

			continue
		}
		mErr.errors = append(mErr.errors, err)
		mErr.mu.Unlock()
	}
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
	mErr.mu.RLock()
	errors := make([]error, len(mErr.errors))
	copy(errors, mErr.errors)
	mErr.mu.RUnlock()

	return errors
}

// ErrOrNil returns MultiError as error, or nil if it does not have
// any stored errors.
func (mErr *MultiError) ErrOrNil() error {
	mErr.mu.RLock()
	defer mErr.mu.RUnlock()

	if len(mErr.errors) == 0 {
		return nil
	}

	return mErr
}

// Format implements fmt.Formatter.
// It relies upon individual error's Format() API if applicable,
// otherwise Error() 's outcome is taken into account.
func (mErr *MultiError) Format(f fmt.State, verb rune) {
	mErr.mu.RLock()
	defer mErr.mu.RUnlock()

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
	mErr.mu.RLock()
	defer mErr.mu.RUnlock()

	if len(mErr.errors) <= 1 {
		return nil
	}

	newMultiErr := NewMultiError()
	newMultiErr.Add(mErr.errors[1:]...)

	return newMultiErr
}

// As implements standard error As() API,
// comparing the first error from stored ones.
func (mErr *MultiError) As(target interface{}) bool {
	mErr.mu.RLock()
	defer mErr.mu.RUnlock()

	return errors.As(mErr.errors[0], target)
}

// Is implements standard error Is() API,
// comparing the first error from stored ones.
func (mErr *MultiError) Is(target error) bool {
	mErr.mu.RLock()
	defer mErr.mu.RUnlock()

	return errors.Is(mErr.errors[0], target)
}
