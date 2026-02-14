// Copyright The ActForGood Authors.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://github.com/actforgood/xerr/blob/main/LICENSE.

package xerr

import (
	"path/filepath"
	"strings"
)

var (
	skipFrame            SkipFrame = AllowFrame
	frameFnNameProcessor FrameFnNameProcessor
)

// SetSkipFrame configures the function this package uses
// in order to include/exclude frames from a stack trace of an error.
// You will call it usually somewhere in the bootstrap process of your
// application. For example:
//
//	// myapp/bootstrap.go
//	func init() {
//		xerr.SetSkipFrame(SkipFoo(SkipBar(xerr.AllowFrame)))
//	}
func SetSkipFrame(fn SkipFrame) {
	skipFrame = fn
}

// SkipFrame is alias for a function that decides whether
// a frame should be included in the stack trace or not.
type SkipFrame func(fnName, file string) bool

// SkipFrameChain is a alias for a chained [SkipFrame].
type SkipFrameChain func(next SkipFrame) SkipFrame

// SkipFrameGoRootSrcPath is a chained function which blacklists
// frames from standard library.
func SkipFrameGoRootSrcPath(next SkipFrame) SkipFrame {
	return func(fnName, file string) bool {
		// Stdlib functions have no module path prefix
		if strings.HasPrefix(fnName, "runtime.") ||
			strings.HasPrefix(fnName, "testing.") ||
			strings.HasPrefix(fnName, "internal/") ||
			strings.Contains(file, string(filepath.Separator)+"go"+string(filepath.Separator)+"src"+string(filepath.Separator)) {
			return true
		}

		// if it's not a stdlib frame, we check next skip frame in the chain.
		return next(fnName, file)
	}
}

// AllowFrame is a [SkipFrame] which whitelists any given frame.
// It can be used as the default/first [SkipFrame] in a chained
// responsibility configuration.
//
// Example:
//
//	xerr.SetSkipFrame(SkipFoo(SkipBar(xerr.AllowFrame)))
func AllowFrame(_, _ string) bool {
	return false
}

// FrameFnNameProcessor is an alias for a function that can
// manipulate the function name from a stack trace frame.
// You can apply customizations upon function name output this way.
type FrameFnNameProcessor func(fnName string) string

// ShortFunctionName is a [FrameFnNameProcessor] which returns only the
// <package.funcName>, removing the fully qualified package name parts.
// Example: "github.com/actforgood/xerr_test.TestX" => "xerr_test.TestX" .
func ShortFunctionName(fnName string) string {
	if lastSlashPos := strings.LastIndex(fnName, "/"); lastSlashPos >= 0 {
		fnName = fnName[lastSlashPos+1:]
	}

	return fnName
}

// OnlyFunctionName is a [FrameFnNameProcessor] which returns only the function
// name, removing the package part.
// Example: "github.com/actforgood/xerr_test.TestX" => "TestX" .
func OnlyFunctionName(fnName string) string {
	fnName = ShortFunctionName(fnName)
	if firstDotPos := strings.Index(fnName, "."); firstDotPos >= 0 {
		fnName = fnName[firstDotPos+1:]
	}

	return fnName
}

// NoDomainFunctionName is a [FrameFnNameProcessor] which removes the first
// part (which is usually a domain) from fully qualified package name.
// Example: "github.com/actforgood/xerr_test.TestX" => "actforgood/xerr_test.TestX" .
func NoDomainFunctionName(fnName string) string {
	if firstSlashPos := strings.Index(fnName, "/"); firstSlashPos >= 0 {
		fnName = fnName[firstSlashPos+1:]
	}

	return fnName
}

// SetFrameFnNameProcessor configures the function this package uses
// in order to manipulate the function name from a stack trace frame.
// You will call it usually somewhere in the bootstrap process of your
// application. For example:
//
//	// myapp/bootstrap.go
//	func init() {
//		xerr.SetFrameFnNameProcessor(xerr.ShortFunctionName)
//	}
func SetFrameFnNameProcessor(fn FrameFnNameProcessor) {
	frameFnNameProcessor = fn
}
