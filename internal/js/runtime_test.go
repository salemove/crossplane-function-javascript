package js

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRuntime_RunScript(t *testing.T) {
	cases := []struct {
		desc         string
		script       string
		args         []interface{}
		transpile    bool
		ok           bool
		expected     interface{}
		expectedJSON string
		expectedYaml string
	}{
		{
			desc:      "no arguments",
			script:    `export default () => { return 11 }`,
			ok:        true,
			transpile: true,
			expected:  11,
		},
		{
			desc:      "function with arguments",
			script:    `export default n => { return n + 1 }`,
			args:      []interface{}{10},
			transpile: true,
			ok:        true,
			expected:  11,
		},
		{
			desc:   "function with complex arguments",
			script: `export default n => { return n.number + 1 }`,
			args: []interface{}{
				map[string]interface{}{
					"number": 10,
				},
			},
			transpile: true,
			ok:        true,
			expected:  11,
		},
		{
			desc:   "struct fields and methods are uncapitalized",
			script: `export default o => { return o.add(o.number, 1) }`,
			args: []interface{}{
				struct {
					Number int
					Add    func(int, int) int
				}{
					Number: 10,
					Add:    func(a int, b int) int { return a + b },
				},
			},
			transpile: true,
			ok:        true,
			expected:  11,
		},
		{
			desc:      "script without exported non-function",
			script:    `export default "str";`,
			ok:        false,
			transpile: true,
		},
		{
			desc:      "script with exported null",
			script:    `export default null;`,
			ok:        false,
			transpile: true,
		},
		{
			desc:      "script without exported default value",
			script:    `export const foo = () => {};`,
			ok:        false,
			transpile: true,
		},
		{
			desc:      "script attempting to trick transpiler",
			script:    `export function _default() { return 1 }`,
			ok:        false,
			transpile: true,
		},
		{
			desc:   "module.exports are not supported",
			script: `"use strict"; module.exports = (n) => { return n + 1 }`,
			ok:     false,
		},
		{
			desc:      "script with syntax errors",
			script:    "!",
			ok:        false,
			transpile: true,
		},
		{
			desc:      "errors in the exported function",
			script:    `export default function() { throw("error") }`,
			ok:        false,
			transpile: true,
		},
		{
			desc:         "JSON.stringify",
			script:       `export default function() { return JSON.stringify({foo: "bar", bar: "baz"}) }`,
			expectedJSON: `{"foo":"bar","bar":"baz"}`,
			ok:           true,
			transpile:    true,
		},
		{
			desc:      "JSON.parse",
			script:    `export default function() { return JSON.parse('{"foo":"bar","bar":"baz"}') }`,
			expected:  map[string]interface{}{"foo": "bar", "bar": "baz"},
			ok:        true,
			transpile: true,
		},
		{
			desc:      "JSON.parse error",
			script:    `export default function() { return JSON.parse('{"foo":') }`,
			ok:        false,
			transpile: true,
		},
		{
			desc:      "btoa",
			script:    `export default function() { return btoa('Hēłłõ, wöřłď') }`,
			expected:  "SMSTxYLFgsO1LCB3w7bFmcWCxI8=",
			ok:        true,
			transpile: true,
		},
		{
			desc:      "atob",
			script:    `export default function() { return atob('SMSTxYLFgsO1LCB3w7bFmcWCxI8=') }`,
			expected:  "Hēłłõ, wöřłď",
			ok:        true,
			transpile: true,
		},
		{
			desc:      "atob error",
			script:    `export default function() { return atob('YWJjZA=') }`,
			ok:        false,
			transpile: true,
		},
		{
			desc:      "without transpile",
			script:    `exports.default = function() { return 1 }`,
			ok:        true,
			transpile: false,
			expected:  1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			r := NewRuntime()
			script := r.Script("test.js", tc.script, tc.args...)
			res, err := script.Run(TranspileToES5(tc.transpile))

			if tc.ok {
				require.NoError(t, err)

				switch {
				case tc.expectedJSON != "":
					assert.JSONEq(t, tc.expectedJSON, res.(string))
				case tc.expectedYaml != "":
					assert.YAMLEq(t, tc.expectedYaml, res.(string))
				default:
					assert.EqualValues(t, tc.expected, res)
				}
			} else {
				require.Error(t, err)
			}
		})
	}
}
