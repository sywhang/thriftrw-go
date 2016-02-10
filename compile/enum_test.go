// Copyright (c) 2015 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package compile

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/thriftrw/thriftrw-go/ast"
	"github.com/thriftrw/thriftrw-go/idl"
	"github.com/thriftrw/thriftrw-go/wire"
)

func i(x int) *int {
	return &x
}

func parseEnum(s string) *ast.Enum {
	prog, err := idl.Parse([]byte(s))
	if err != nil {
		panic(fmt.Sprintf("failure to parse: %v: %s", err, s))
	}

	if len(prog.Definitions) != 1 {
		panic("parseEnum may be used to parse single enums only")
	}

	return prog.Definitions[0].(*ast.Enum)
}

func TestCompileEnumSuccess(t *testing.T) {
	tests := []struct {
		src  string
		spec *EnumSpec
	}{
		{
			// Default values
			"enum Role { Disabled, User, Moderator, Admin }",
			&EnumSpec{
				Name: "Role",
				Items: []EnumItem{
					EnumItem{"Disabled", 0},
					EnumItem{"User", 1},
					EnumItem{"Moderator", 2},
					EnumItem{"Admin", 3},
				},
			},
		},
		{
			// Explicit values
			"enum CommentStatus { Visible = 12345, Hidden = 54321 }",
			&EnumSpec{
				Name: "CommentStatus",
				Items: []EnumItem{
					EnumItem{"Visible", 12345},
					EnumItem{"Hidden", 54321},
				},
			},
		},
		{
			// Mixed
			"enum foo { A, B, C = 10, D, E }",
			&EnumSpec{
				Name: "foo",
				Items: []EnumItem{
					EnumItem{"A", 0},
					EnumItem{"B", 1},
					EnumItem{"C", 10},
					EnumItem{"D", 11},
					EnumItem{"E", 12},
				},
			},
		},
		{
			// Same values
			"enum bar { A, B = 0, C, D = 0, E }",
			&EnumSpec{
				Name: "bar",
				Items: []EnumItem{
					EnumItem{"A", 0},
					EnumItem{"B", 0},
					EnumItem{"C", 1},
					EnumItem{"D", 0},
					EnumItem{"E", 1},
				},
			},
		},
	}

	for _, tt := range tests {
		src := parseEnum(tt.src)
		enumspec, err := compileEnum(src)
		if assert.NoError(t, err) {
			spec, err := enumspec.Link(scope())
			assert.NoError(t, err)
			assert.Equal(t, wire.TI32, spec.TypeCode())
			assert.Equal(t, tt.spec, spec)

			// compiling twice should not error
			spec, err = spec.Link(scope())
			assert.NoError(t, err)
		}
	}
}

func TestCompileEnumFailure(t *testing.T) {
	tests := []struct {
		src      string
		messages []string
	}{
		{
			"enum Foo { A, B, C, A, D }",
			[]string{
				`cannot compile "Foo.A"`,
				`the name "A" has already been used`,
			},
		},
	}

	for _, tt := range tests {
		src := parseEnum(tt.src)
		_, err := compileEnum(src)

		if assert.Error(t, err) {
			for _, msg := range tt.messages {
				assert.Contains(t, err.Error(), msg)
			}
		}
	}
}
