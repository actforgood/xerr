// Copyright The ActForGood Authors.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://github.com/actforgood/xerr/blob/main/LICENSE.

package xerr_test

import (
	"os"
	"runtime"
	"testing"

	"github.com/actforgood/xerr"
)

func TestAllowFrame(t *testing.T) {
	t.Parallel()

	// arrange
	subject := xerr.AllowFrame

	for range 10 {
		// act
		result := subject("foo", "bar")

		// assert
		assertFalse(t, result)
	}
}

func TestSkipFrameGoRootSrcPath(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		subject      = xerr.SkipFrameGoRootSrcPath
		nextCallsCnt = 0
		tests        = [...]struct {
			name      string
			inputFile string
			next      xerr.SkipFrame
			expected  bool
		}{
			{
				name:      "random path, expect false",
				inputFile: "/foo/bar/baz.go",
				next:      xerr.AllowFrame,
				expected:  false,
			},
			{
				name:      "random path, with next that skips frame, expect true",
				inputFile: "/foo/bar/baz.go",
				next: func(_, _ string) bool {
					nextCallsCnt++

					return true
				},
				expected: true,
			},
			{
				name:      "GOROOT/bin, expect false",
				inputFile: runtime.GOROOT() + string(os.PathSeparator) + "bin/foo",
				next:      xerr.AllowFrame,
				expected:  false,
			},
			{
				name:      "GOROOT/src, expect true",
				inputFile: runtime.GOROOT() + string(os.PathSeparator) + "src/foo/bar.go",
				next:      xerr.AllowFrame,
				expected:  true,
			},
		}
	)

	for _, testData := range tests {
		test := testData // capture range variable
		t.Run(test.name, func(t *testing.T) {
			// act
			result := subject(test.next)("runtime.goexit", test.inputFile)

			// assert
			assertEqual(t, test.expected, result)
		})
	}
	assertEqual(t, 1, nextCallsCnt)
}

func TestShortFunctionName(t *testing.T) {
	t.Parallel()

	// arrange
	subject := xerr.ShortFunctionName
	tests := [...]struct {
		name        string
		inputFnName string
		expected    string
	}{
		{
			name:        "empty, expect empty",
			inputFnName: "",
			expected:    "",
		},
		{
			name:        "random string, expect same string",
			inputFnName: "Function",
			expected:    "Function",
		},
		{
			name:        "short function name, expect same string",
			inputFnName: "pkg.Function",
			expected:    "pkg.Function",
		},
		{
			name:        "short pointer method name, expect same string",
			inputFnName: "pkg.(*Class).Method",
			expected:    "pkg.(*Class).Method",
		},
		{
			name:        "simple fully package qualified function name, expect short function name",
			inputFnName: "example.com/foo/pkg.Function",
			expected:    "pkg.Function",
		},
		{
			name:        "simple fully package qualified pointer method name, expect short function name",
			inputFnName: "example.com/foo/pkg.(*Class).Method",
			expected:    "pkg.(*Class).Method",
		},
		{
			name:        "fully sub-package qualified function name, expect short function name",
			inputFnName: "github.com/actforgood/xerr/subpkg.Function",
			expected:    "subpkg.Function",
		},
		{
			name:        "fully sub-package qualified pointer method name, expect short function name",
			inputFnName: "github.com/actforgood/xerr/subpkg.(*Class).Function",
			expected:    "subpkg.(*Class).Function",
		},
	}

	for _, testData := range tests {
		test := testData // capture range variable
		t.Run(test.name, func(t *testing.T) {
			// act
			result := subject(test.inputFnName)

			// assert
			assertEqual(t, test.expected, result)
		})
	}
}

func TestOnlyFunctionName(t *testing.T) {
	t.Parallel()

	// arrange
	subject := xerr.OnlyFunctionName
	tests := [...]struct {
		name        string
		inputFnName string
		expected    string
	}{
		{
			name:        "empty, expect empty",
			inputFnName: "",
			expected:    "",
		},
		{
			name:        "random string, expect same string",
			inputFnName: "Function",
			expected:    "Function",
		},
		{
			name:        "short function name, expect only func name",
			inputFnName: "pkg.Function",
			expected:    "Function",
		},
		{
			name:        "short pointer method name, expect only func name",
			inputFnName: "pkg.(*Class).Method",
			expected:    "(*Class).Method",
		},
		{
			name:        "simple fully package qualified function name, expect only func name",
			inputFnName: "example.com/foo/pkg.Function",
			expected:    "Function",
		},
		{
			name:        "simple fully package qualified pointer method name, expect only func name",
			inputFnName: "example.com/foo/pkg.(*Class).Method",
			expected:    "(*Class).Method",
		},
		{
			name:        "fully sub-package qualified function name, expect short function name",
			inputFnName: "github.com/actforgood/xerr/subpkg.Function",
			expected:    "Function",
		},
		{
			name:        "fully sub-package qualified pointer method name, expect only func name",
			inputFnName: "github.com/actforgood/xerr/subpkg.(*Class).Function",
			expected:    "(*Class).Function",
		},
	}

	for _, testData := range tests {
		test := testData // capture range variable
		t.Run(test.name, func(t *testing.T) {
			// act
			result := subject(test.inputFnName)

			// assert
			assertEqual(t, test.expected, result)
		})
	}
}

func TestNoDomainFunctionName(t *testing.T) {
	t.Parallel()

	// arrange
	subject := xerr.NoDomainFunctionName
	tests := [...]struct {
		name        string
		inputFnName string
		expected    string
	}{
		{
			name:        "empty, expect empty",
			inputFnName: "",
			expected:    "",
		},
		{
			name:        "random string, expect same string",
			inputFnName: "Function",
			expected:    "Function",
		},
		{
			name:        "short function name, expect same string",
			inputFnName: "pkg.Function",
			expected:    "pkg.Function",
		},
		{
			name:        "short pointer method name, expect same string",
			inputFnName: "pkg.(*Class).Method",
			expected:    "pkg.(*Class).Method",
		},
		{
			name:        "simple fully package qualified function name, expect no domain function name",
			inputFnName: "example.com/foo/pkg.Function",
			expected:    "foo/pkg.Function",
		},
		{
			name:        "simple fully package qualified pointer method name, expect short function name",
			inputFnName: "example.com/foo/pkg.(*Class).Method",
			expected:    "foo/pkg.(*Class).Method",
		},
		{
			name:        "fully sub-package qualified function name, expect short function name",
			inputFnName: "github.com/actforgood/xerr/subpkg.Function",
			expected:    "actforgood/xerr/subpkg.Function",
		},
		{
			name:        "fully sub-package qualified pointer method name, expect short function name",
			inputFnName: "github.com/actforgood/xerr/subpkg.(*Class).Function",
			expected:    "actforgood/xerr/subpkg.(*Class).Function",
		},
	}

	for _, testData := range tests {
		test := testData // capture range variable
		t.Run(test.name, func(t *testing.T) {
			// act
			result := subject(test.inputFnName)

			// assert
			assertEqual(t, test.expected, result)
		})
	}
}
