// Copyright 2022 Bogdan Constantinescu.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://github.com/actforgood/xerr/LICENSE.

package xerr_test

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/actforgood/xerr"
)

func ExampleNew() {
	err := xerr.New("something went bad")

	fmt.Println(err.Error())
	// Example output:
	// something went bad

	fmt.Printf("%+v\n", err)
	// Example output:
	// something went bad
	// github.com/actforgood/xerr_test.ExampleNew
	// 		/Users/bogdan/work/go/xerr/example_test.go:15
	// testing.runExample
	// 		/usr/local/go/src/testing/run_example.go:63
	// testing.runExamples
	//		/usr/local/go/src/testing/example.go:44
	// testing.(*M).Run
	//		/usr/local/go/src/testing/testing.go:1418
	// main.main
	//		_testmain.go:79
	// runtime.main
	// 		/usr/local/go/src/runtime/proc.go:225
	// runtime.goexit
	// 		/usr/local/go/src/runtime/asm_amd64.s:1371
}

func ExampleErrorf() {
	err := xerr.Errorf("something went bad with %s", "this example")

	fmt.Println(err.Error())
	// Example output:
	// something went bad with this example

	fmt.Printf("%+v\n", err)
	// Example output:
	// something went bad with this example
	// github.com/actforgood/xerr_test.ExampleErrorf
	// 		/Users/bogdan/work/go/xerr/example_test.go:44
	// testing.runExample
	// 		/usr/local/go/src/testing/run_example.go:63
	// testing.runExamples
	//		/usr/local/go/src/testing/example.go:44
	// testing.(*M).Run
	//		/usr/local/go/src/testing/testing.go:1418
	// main.main
	//		_testmain.go:83
	// runtime.main
	// 		/usr/local/go/src/runtime/proc.go:225
	// runtime.goexit
	// 		/usr/local/go/src/runtime/asm_amd64.s:1371
}

// DoSomeOperation simulates a function which (may) return an error.
func DoSomeOperation() error {
	return errors.New("op err")
}

// DoSomeOtherOperation simulates a function which (may) return an error.
func DoSomeOtherOperation() error {
	return xerr.New("op err")
}

func ExampleWrap_withStandardError() {
	err := DoSomeOperation()
	if err != nil {
		err = xerr.Wrap(err, "could not perform operation")
	}

	fmt.Println(err.Error())
	// Example output:
	// could not perform operation: op err

	fmt.Printf("%+v\n", err)
	// Example output:
	// could not perform operation: op err
	// github.com/actforgood/xerr_test.ExampleWrap_withStandardError
	// 		/Users/bogdan/work/go/xerr/example_test.go:54
	// testing.runExample
	// 		/usr/local/go/src/testing/run_example.go:63
	// testing.runExamples
	//		/usr/local/go/src/testing/example.go:44
	// testing.(*M).Run
	//		/usr/local/go/src/testing/testing.go:1418
	// main.main
	//		_testmain.go:81
	// runtime.main
	// 		/usr/local/go/src/runtime/proc.go:225
	// runtime.goexit
	// 		/usr/local/go/src/runtime/asm_amd64.s:1371
}

func ExampleWrap_withStackError() {
	err := DoSomeOtherOperation()
	if err != nil {
		err = xerr.Wrap(err, "could not perform operation")
	}

	fmt.Println(err.Error())
	// Example output:
	// could not perform operation: op err

	fmt.Printf("%+v\n", err)
	// Example output:
	// could not perform operation: op err
	// github.com/actforgood/xerr_test.ExampleWrap_withStackError
	//         /Users/bogdan/work/go/xerr/example_test.go:83
	// github.com/actforgood/xerr_test.DoSomeOtherOperation
	//         /Users/bogdan/work/go/xerr/example_test.go:48
	// github.com/actforgood/xerr_test.ExampleWrap_withStackError
	//         /Users/bogdan/work/go/xerr/example_test.go:81
	// testing.runExample
	//         /usr/local/go/src/testing/run_example.go:63
	// testing.runExamples
	//         /usr/local/go/src/testing/example.go:44
	// testing.(*M).Run
	//         /usr/local/go/src/testing/testing.go:1418
	// main.main
	//         _testmain.go:79
	// runtime.main
	//         /usr/local/go/src/runtime/proc.go:225
	// runtime.goexit
	//         /usr/local/go/src/runtime/asm_amd64.s:1371
}

func ExampleMultiError_Add_sequential() {
	files := []string{
		"/this/file/does/not/exist/1",
		"/this/file/does/not/exist/2",
		"/this/file/does/not/exist/3",
	}
	var multiErr *xerr.MultiError // save allocation if Add is never called.
	for _, file := range files {
		if _, err := os.Open(file); err != nil {
			multiErr = multiErr.Add(err) // store allocated multiErr.
		}
		// else do something with that file ...
	}

	returnErr := multiErr.ErrOrNil()
	fmt.Println(returnErr)

	// Example of output: note - on Windows the message is slightly different (The system cannot find the path specified)
	// error #1
	// open /this/file/does/not/exist/1: no such file or directory
	// error #2
	// open /this/file/does/not/exist/2: no such file or directory
	// error #3
	// open /this/file/does/not/exist/3: no such file or directory
}

func ExampleMultiError_Add_parallel() {
	files := []string{
		"/this/file/does/not/exist/1",
		"/this/file/does/not/exist/2",
		"/this/file/does/not/exist/3",
	}
	multiErr := xerr.NewMultiError() // we need instance already initialized.
	var wg sync.WaitGroup
	for _, file := range files {
		wg.Add(1)
		go func(filePath string, mErr *xerr.MultiError, waitGr *sync.WaitGroup) {
			defer waitGr.Done()
			if _, err := os.Open(filePath); err != nil {
				_ = mErr.Add(err) // we can dismiss returned value as multiErr is already initialized.
			}
			// else do something with that file ...
		}(file, multiErr, &wg)
	}
	wg.Wait()

	returnErr := multiErr.ErrOrNil()
	fmt.Println(returnErr)

	// Example of unordered output:
	// error #1
	// open /this/file/does/not/exist/3: no such file or directory
	// error #2
	// open /this/file/does/not/exist/1: no such file or directory
	// error #3
	// open /this/file/does/not/exist/2: no such file or directory
}

func ExampleMultiError_AddOnce() {
	// different errors we want to add to MultiError
	err1 := os.ErrNotExist
	err2 := io.ErrUnexpectedEOF
	err3 := os.ErrNotExist

	var multiErr = xerr.NewMultiError()
	_ = multiErr.AddOnce(err1)
	_ = multiErr.AddOnce(err2)
	_ = multiErr.AddOnce(err3) // err3 is the same with err1, so it should be ignored

	returnErr := multiErr.ErrOrNil()
	fmt.Println(returnErr)

	// Output:
	// error #1
	// file does not exist
	// error #2
	// unexpected EOF
}

func ExampleMultiError_Errors() {
	var multiErr = xerr.NewMultiError()
	_ = multiErr.Add(errors.New("1st error"))
	_ = multiErr.Add(errors.New("2nd error"))

	for _, err := range multiErr.Errors() {
		fmt.Println(err)
	}

	// Output:
	// 1st error
	// 2nd error
}

func ExampleMultiError_Is() {
	var multiErr = xerr.NewMultiError()
	_ = multiErr.Add(io.ErrUnexpectedEOF)
	someErrWithStack := xerr.New("stack err")
	_ = multiErr.Add(someErrWithStack)

	fmt.Println(errors.Is(multiErr, io.ErrUnexpectedEOF))
	fmt.Println(errors.Is(multiErr, someErrWithStack))
	fmt.Println(errors.Is(multiErr, os.ErrClosed))

	// Output:
	// true
	// true
	// false
}
