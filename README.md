# Xerr

[![Build Status](https://github.com/actforgood/xerr/actions/workflows/build.yml/badge.svg)](https://github.com/actforgood/xerr/actions/workflows/build.yml)
[![License](https://img.shields.io/badge/license-MIT-blue)](https://raw.githubusercontent.com/actforgood/xerr/main/LICENSE)
[![Coverage Status](https://coveralls.io/repos/github/actforgood/xerr/badge.svg?branch=main)](https://coveralls.io/github/actforgood/xerr?branch=main)
[![Go Reference](https://pkg.go.dev/badge/github.com/actforgood/xerr.svg)](https://pkg.go.dev/github.com/actforgood/xerr)  

---

Package `xerr` provides an error with stack trace, a multi-error and other different types of errors.  


### Features:
* an error enriched with stack trace
* a MultiError


### Error with stack trace
Basic example:  
```golang
// create a new error with stack trace and print it.
err := xerr.New("something went bad")
fmt.Printf("%+v", err)
```
Output example:
```
something went bad
github.com/actforgood/xerr/_example/pkgb.OperationB
    /Users/bogdan/work/go/xerr/_example/pkgb/otherfile.go:7
github.com/actforgood/xerr/_example/pkga.OperationA
    /Users/bogdan/work/go/xerr/_example/pkga/somefile.go:6
main.main
    /Users/bogdan/work/go/xerr/_example/main.go:14
runtime.main
    /usr/local/go/src/runtime/proc.go:225
runtime.goexit
    /usr/local/go/src/runtime/asm_amd64.s:1371
```

Wrap another error example:
```
err := DoSomeOperation()
if err != nil {
    err = xerr.Wrap(err, "could not perform operation")
}
```
Output example:
```
could not perform operation: op err
github.com/actforgood/xerr/_example/pkgb.OperationB
    /Users/bogdan/work/go/xerr/_example/pkgb/otherfile.go:14
github.com/actforgood/xerr/_example/pkga.OperationA
    /Users/bogdan/work/go/xerr/_example/pkga/somefile.go:6
main.main
    /Users/bogdan/work/go/xerr/_example/main.go:14
runtime.main
    /usr/local/go/src/runtime/proc.go:225
runtime.goexit
    /usr/local/go/src/runtime/asm_amd64.s:1371
```

##### Shrinking the size of your error's output
You can reduce the I/O bytes and/or storage for your (logged) errors by shrinking the output of stack traces.  
The package provides ways of manipulating the function name and excluding frames from the stack trace. 
- Example of excluding frames like /usr/local/go/src/ (which is my GOROOT src path):
```
// somewhere in your application bootstrap:
func init() {
    xerr.SetSkipFrame(xerr.SkipFrameGoRootSrcPath(xerr.AllowFrame))
}
```
Let's see how error's output looks like now:
```
something went bad
github.com/actforgood/xerr/_example/pkgb.OperationB
    /Users/bogdan/work/go/xerr/_example/pkgb/otherfile.go:11
github.com/actforgood/xerr/_example/pkga.OperationA
    /Users/bogdan/work/go/xerr/_example/pkga/somefile.go:6
main.main
    /Users/bogdan/work/go/xerr/_example/main.go:15
```
You can implement other rules of exclusion by yourself, and even chain multiple rules. Check `SkipFrame` and `SkipFrameChain`.
- Example of saving some bytes by shorting the function name.
```
// somewhere in your application bootstrap:
func init() {
    xerr.SetFrameFnNameProcessor(xerr.ShortFunctionName)
}
```
Let's see how error's output looks like now:
```
something went bad
pkgb.OperationB
    /Users/bogdan/work/go/xerr/_example/pkgb/otherfile.go:11
pkga.OperationA
    /Users/bogdan/work/go/xerr/_example/pkga/somefile.go:6
main.main
    /Users/bogdan/work/go/xerr/_example/main.go:16
```
Check also other function names shrinkers: `OnlyFunctionName`, `NoDomainFunctionName`.
- Tip: you can also shrink the filenames. This is not covered by this pkg, but you can achieve it by a go build/run flag.
You can read [this](https://itnext.io/trim-gopath-from-stack-trace-88b7402c8b47) article.
```
go run -gcflags "all=-trimpath=/Users/bogdan/work/go" /Users/bogdan/work/go/xerr/_example/main.go
something went bad
pkgb.OperationB
    xerr/_example/pkgb/otherfile.go:11
pkga.OperationA
    xerr/_example/pkga/somefile.go:6
main.main
    xerr/_example/main.go:16
```



### MultiError
You can collect multiple errors into a `MultiError` which implements `error` interface.
Basic sequential example:
```golang
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

err := multiErr.ErrOrNil()
fmt.Println(err)
return err
```
Output example:
```
error #1
open /this/file/does/not/exist/1: no such file or directory
error #2
open /this/file/does/not/exist/2: no such file or directory
error #3
open /this/file/does/not/exist/3: no such file or directory
```

Basic parallel example:
```golang
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
```
Output example:
```
error #1
open /this/file/does/not/exist/3: no such file or directory
error #2
open /this/file/does/not/exist/1: no such file or directory
error #3
open /this/file/does/not/exist/2: no such file or directory
```    

### Misc 
Feel free to use this pkg if you like it and fits your needs.  
Check also other stack aware errors packages like pkg\errors, go-errors\errors.  
For multi-error there are hashicorp\go-multierror, uber-go\multierr.
Here stands some benchmarks made locally for stacked error (note though each package err output may be different):  
```
goos: darwin
goarch: amd64
pkg: github.com/actforgood/xerr
cpu: Intel(R) Core(TM) i7-7700HQ CPU @ 2.80GHz
BenchmarkNewPkgErrors-8          1409167              4396 ns/op             680 B/op         11 allocs/op
BenchmarkNewGoErrors-8           48172              126978 ns/op           21255 B/op         71 allocs/op
BenchmarkNewXerr-8               2409250              2495 ns/op             656 B/op          7 allocs/op
```
```golang
// code snippet for other packages bench
import (
    "fmt"
    "testing"

    goErr "github.com/go-errors/errors"
    pkgErr "github.com/pkg/errors"
)

func BenchmarkNewPkgErrors(b *testing.B) {
    for n := 0; n < b.N; n++ {
        err := pkgErr.New("some error with stack trace")
        _ = fmt.Sprintf("%+v", err)
    }
}

func BenchmarkNewGoErrors(b *testing.B) {
    for n := 0; n < b.N; n++ {
        err := goErr.New("some error with stack trace")
        _ = err.ErrorStack()
    }
}
```


### License
This package is released under a MIT license. See [LICENSE](LICENSE).  
