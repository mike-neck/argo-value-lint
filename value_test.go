package main

import (
	"reflect"
	"testing"
)

func TestParseValue(t *testing.T) {
	all := []struct {
		title    string
		value    string
		expected Value
	}{
		{
			title: "plain text should be a RawValue",
			value: "test-value",
			expected: Value{
				{RawValue, "test-value"},
			},
		},
		{
			title: "single expression should be an Variable",
			value: "{{item}}",
			expected: Value{
				{Variable, "item"},
			},
		},
		{
			title: "multiple expressions should be a combination of RawValue and Variable",
			value: "{{item1}}-test-{{item2}}",
			expected: Value{
				{Variable, "item1"},
				{RawValue, "-test-"},
				{Variable, "item2"},
			},
		},
		{
			title: "invalid expression '{item}}' should be a RawValue",
			value: "{item}}",
			expected: Value{
				{RawValue, "{item}}"},
			},
		},
		{
			title: "invalid expression '{{item}' should be a RawValue",
			value: "{{item}",
			expected: Value{
				{RawValue, "{{item}"},
			},
		},
		{
			title: "invalid expression '{{item' should be a RawValue",
			value: "{{item",
			expected: Value{
				{RawValue, "{{item"},
			},
		},
		{
			title: "an expression which has parenthesis pair should be a Variable",
			value: "{{this{test}expression}}",
			expected: Value{
				{Variable, "this{test}expression"},
			},
		},
		{
			title: "unmatched parenthesis should be a RawValue",
			value: "value}}value",
			expected: Value{
				{RawValue, "value}}value"},
			},
		},
		{
			title: "a json should be a RawValue",
			value: `{"type": "json", "value": "test-value"}`,
			expected: Value{
				{RawValue, `{"type": "json", "value": "test-value"}`},
			},
		},
		{
			title: "an escaped parenthesis should be a RawValue",
			value: `\{\{item\}\}`,
			expected: Value{
				{RawValue, `{{item}}`},
			},
		},
		{
			title: "an escaped n should be a RawValue with new line",
			value: `aaa\nbbb`,
			expected: Value{
				{RawValue, `aaa
bbb`},
			},
		},
		{
			title: "an expression starting with '=' should be a Expression",
			value: "{{=item.value}}",
			expected: Value{
				{Expression, "item.value"},
			},
		},
	}
	for _, v := range all {
		t.Run(v.title, func(t *testing.T) {
			value := ParseValue(v.value)
			if !reflect.DeepEqual(value, v.expected) {
				t.Errorf("\nExpected: %v\nGot: %v", v.expected, value)
			}
		})
	}
}
