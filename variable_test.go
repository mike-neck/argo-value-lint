package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"path/filepath"
	"strings"
	"testing"

	testdataloader "github.com/peteole/testdata-loader"
)

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

func TestVariableNode_NestedCallMultipleTimes(t *testing.T) {
	root := NewVariableNode()
	root.Nested("workflow", func(workflow *VariableNode) {
		workflow.Next("test")
	})
	root.Nested("workflow", func(workflow *VariableNode) {
		workflow.Next("example")
	})
	root.Nested("steps", func(steps *VariableNode) {
		steps.PatternNested(LowerCaseWithHyphen, func(step *VariableNode) {
			step.Next("id")
		})
	})

	tests := []struct {
		name string
		pass bool
	}{
		{
			name: "workflow.test",
			pass: true,
		},
		{
			name: "workflow.example",
			pass: true,
		},
		{
			name: "input.test",
			pass: false,
		},
		{
			name: "steps.example.id",
			pass: true,
		},
		{
			name: "steps.example.value",
			pass: false,
		},
		{
			name: "steps.example",
			pass: false,
		},
	}
	for i, test := range tests {
		id := i + 1
		t.Run(fmt.Sprintf("%d %s", id, test.name), func(t *testing.T) {
			fragments := strings.Split(test.name, ".")
			err := root.Validate(0, fragments)
			if test.pass && err != nil {
				t.Errorf("got unexpected error: %s", err)
			} else if !test.pass && err == nil {
				t.Errorf("expected error but got nil")
			}
		})
	}
}

func TestNewValidator(t *testing.T) {
	categories := []struct {
		name string
		expr ExpressionContext
	}{
		{"workflow", GlobalExpression},
		{"inputs", AllTemplates},
		{"steps", StepsTemplates},
		{"dag", DAGTemplates},
		{"http", HTTPTemplates},
	}
	for _, category := range categories {
		t.Run(category.name, func(t *testing.T) {
			validator := NewValidator(category.expr)
			testFile := filepath.Join("test", "variable", fmt.Sprintf("%s.yaml", category.name))
			data := testdataloader.GetTestFile(testFile)
			var tests struct {
				Tests []struct {
					Name string `yaml:"name"`
					Pass bool   `yaml:"pass"`
				} `yaml:"tests"`
			}
			err := yaml.Unmarshal(data, &tests)
			if err != nil {
				t.Fatalf("could not unmarshal test data: %s", err)
				return
			} else if len(tests.Tests) == 0 {
				t.Fatalf("no Tests found")
				return
			}
			t.Log("validator", "\n", validator.debugDescription(0))
			for _, test := range tests.Tests {
				t.Run(test.Name, func(t *testing.T) {
					fragments := strings.Split(test.Name, ".")
					err := validator.Validate(0, fragments)
					if test.Pass && err != nil {
						t.Errorf("got unexpected error: %s", err)
					} else if !test.Pass && err == nil {
						t.Errorf("expected error but got nil")
					}
				})
			}
		})
	}
}
