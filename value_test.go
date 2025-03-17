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
			title: "single expression should be an Expression",
			value: "{{item}}",
			expected: Value{
				{Expression, "item"},
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
