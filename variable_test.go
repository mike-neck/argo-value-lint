package main

import "testing"

func TestStepLoop_Validate(t *testing.T) {
	tests := []struct {
		title       string
		sut         StepLoop
		inputText   string
		expectError bool
	}{
		{
			title:       "text not 'item' and withItem then error",
			sut:         loopWithItems(),
			inputText:   "workflows.name",
			expectError: true,
		},
		{
			title:       "text to be 'item' and withItem then valid",
			sut:         loopWithItems(),
			inputText:   "item",
			expectError: false,
		},
		{
			title:       "text not starting 'item' and withParams then error",
			sut:         loopWithParams("test"),
			inputText:   "workflows.name",
			expectError: true,
		},
		{
			title:       "text to be 'item' and withParams then error",
			sut:         loopWithParams("test"),
			inputText:   "item",
			expectError: true,
		},
		{
			title:       "text with attribute 'item.test' and withParams then valid",
			sut:         loopWithParams("test"),
			inputText:   "item.test",
			expectError: false,
		},
		{
			title:       "text with unknown attribute and withParams then error",
			sut:         loopWithParams("test"),
			inputText:   "item.unknown",
			expectError: true,
		},
	}
	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			err := test.sut.Validate(test.inputText)
			if test.expectError && err == nil {
				t.Errorf("expected error but got nil")
			} else if !test.expectError && err != nil {
				t.Errorf("expected no error but got %s", err)
			}
		})
	}
}
