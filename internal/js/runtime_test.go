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
		ok           bool
		expected     interface{}
		expectedJSON string
		expectedYaml string
	}{
		{
			desc:     "no arguments",
			script:   `export default () => { return 11 }`,
			ok:       true,
			expected: 11,
		},
		{
			desc:     "function with arguments",
			script:   `export default n => { return n + 1 }`,
			args:     []interface{}{10},
			ok:       true,
			expected: 11,
		},
		{
			desc:   "function with complex arguments",
			script: `export default n => { return n.number + 1 }`,
			args: []interface{}{
				map[string]interface{}{
					"number": 10,
				},
			},
			ok:       true,
			expected: 11,
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
			ok:       true,
			expected: 11,
		},
		{
			desc:   "script without exported non-function",
			script: `export default "str";`,
			ok:     false,
		},
		{
			desc:   "script with exported null",
			script: `export default null;`,
			ok:     false,
		},
		{
			desc:   "script without exported default value",
			script: `export const foo = () => {};`,
			ok:     false,
		},
		{
			desc:   "script attempting to trick transpiler",
			script: `export function _default() { return 1 }`,
			ok:     false,
		},
		{
			desc:   "module.exports are not supported",
			script: `"use strict"; module.exports = (n) => { return n + 1 }`,
			ok:     false,
		},
		{
			desc:   "script with syntax errors",
			script: "!",
			ok:     false,
		},
		{
			desc:   "errors in the exported function",
			script: `export default function() { throw("error") }`,
			ok:     false,
		},
		{
			desc:         "JSON.stringify",
			script:       `export default function() { return JSON.stringify({foo: "bar", bar: "baz"}) }`,
			expectedJSON: `{"foo":"bar","bar":"baz"}`,
			ok:           true,
		},
		{
			desc:     "JSON.parse",
			script:   `export default function() { return JSON.parse('{"foo":"bar","bar":"baz"}') }`,
			expected: map[string]interface{}{"foo": "bar", "bar": "baz"},
			ok:       true,
		},
		{
			desc:   "JSON.parse error",
			script: `export default function() { return JSON.parse('{"foo":') }`,
			ok:     false,
		},
		{
			desc:         "YAML.stringify",
			script:       `import * as YAML from "yaml"; export default function() { return YAML.stringify({foo: "bar", "bar": "baz"}) }`,
			expectedYaml: "foo: bar\nbar: baz\n",
			ok:           true,
		},
		{
			desc:     "YAML.parse",
			script:   `import * as YAML from "yaml"; export default function() { return YAML.parse("foo: bar\nbar: baz\n") }`,
			expected: map[string]interface{}{"foo": "bar", "bar": "baz"},
			ok:       true,
		},
		{
			desc:   "YAML.parse error",
			script: `import * as YAML from "yaml"; export default function() { return YAML.parse("foo: bar\n  bar: baz\n") }`,
			ok:     false,
		},
		{
			desc:     "Base64.encode",
			script:   `import * as Base64 from "base64"; export default function() { return Base64.encode('abcd') }`,
			expected: "YWJjZA==",
			ok:       true,
		},
		{
			desc:     "Base64.decode",
			script:   `import * as Base64 from "base64"; export default function() { return Base64.decode('YWJjZA==') }`,
			expected: "abcd",
			ok:       true,
		},
		{
			desc:   "Base64.decode error",
			script: `import * as Base64 from "base64"; export default function() { return Base64.decode('YWJjZA=') }`,
			ok:     false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			r := NewRuntime()

			res, err := r.RunScript("test.js", tc.script, tc.args...)

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
