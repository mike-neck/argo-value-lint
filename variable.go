package main

import (
	"fmt"
	"regexp"
	"strings"
)

type StepLoop struct {
	Children []string
}

func loopWithItems() StepLoop {
	return StepLoop{}
}

func loopWithParams(attributes ...string) StepLoop {
	return StepLoop{
		Children: attributes,
	}
}

func (s *StepLoop) Validate(text string) error {
	if len(s.Children) == 0 {
		if text == "item" {
			return nil
		}
		return fmt.Errorf("withItem cannot access attributes: %s", text)
	}
	if !strings.HasPrefix(text, "item.") {
		return fmt.Errorf("invalid loop variable: %s", text)
	}
	child := strings.Replace(text, "item.", "", 1)
	for _, n := range s.Children {
		if strings.HasPrefix(child, n) {
			return nil
		}
	}
	return fmt.Errorf("withParams has no attribute: %s", text)
}

type ObjectValidation interface {
	Validate(index int, fragments []string) error
}

type VariableNode struct {
	Names    []string
	Children map[string]ObjectValidation
}

func NewVariableNode() *VariableNode {
	names := make([]string, 0)
	children := make(map[string]ObjectValidation)
	return &VariableNode{
		Names:    names,
		Children: children,
	}
}

type EvaluationContext func(*VariableNode)

var (
	EvaluationContexts = struct {
		Workflow EvaluationContext
	}{
		Workflow: workflowValidation,
	}
)

func (v *VariableNode) Validate(index int, fragments []string) error {
	m := len(fragments)
	if m <= index {
		return fmt.Errorf("invalid index: %d", index)
	}
	fragment := fragments[index]
	if index+1 == m {
		for _, name := range v.Names {
			if fragment == name {
				return nil
			}
		}
		return fmt.Errorf("variable not found: %s", strings.Join(fragments, "."))
	}
	for name, validation := range v.Children {
		if fragment == name {
			return validation.Validate(index+1, fragments)
		}
	}
	return fmt.Errorf("variable not found: %s", strings.Join(fragments, "."))
}

func (v *VariableNode) Next(name string) {
	v.Names = append(v.Names, name)
}

func (v *VariableNode) Nested(name string, config func(variableNode *VariableNode)) {
	ov, ok := v.Children[name]
	if !ok {
		children := NewVariableNode()
		config(children)
		v.Children[name] = children
	} else if node, ok := ov.(*VariableNode); ok {
		config(node)
	}
}

func (v *VariableNode) Pattern(name string, patterns ...*regexp.Regexp) {
	var child ObjectValidation
	m := len(patterns)
	if m == 0 {
		return
	}
	for i := m - 1; 0 <= i; i-- {
		pattern := patterns[i]
		if i == m-1 {
			child = &PatternedVariableNameLeafValidator{pattern}
		} else {
			child = &PatternedVariableNameNodeValidator{
				Pattern: pattern,
				node:    child,
			}
		}
	}
	v.Children[name] = child
}

func (v *VariableNode) PatternNested(name string, pattern *regexp.Regexp, config func(variableNode *VariableNode)) {

}

type PatternedVariableNameNodeValidator struct {
	Pattern *regexp.Regexp
	node    ObjectValidation
}

func (p *PatternedVariableNameNodeValidator) Validate(index int, fragments []string) error {
	m := len(fragments)
	if m <= index {
		return fmt.Errorf("invalid index: %d", index)
	}
	if index+1 == m {
		return fmt.Errorf("variable not found: %s", strings.Join(fragments, "."))
	}
	fragment := fragments[index]
	if p.Pattern.Match([]byte(fragment)) {
		return p.node.Validate(index+1, fragments)
	}
	return fmt.Errorf("variable not found: %s", strings.Join(fragments, "."))
}

type PatternedVariableNameLeafValidator struct {
	Pattern *regexp.Regexp
}

func (p *PatternedVariableNameLeafValidator) Validate(index int, fragments []string) error {
	m := len(fragments)
	if index+1 != m {
		return fmt.Errorf("variable not found: %s", strings.Join(fragments, "."))
	}
	fragment := fragments[index]
	if p.Pattern.Match([]byte(fragment)) {
		return nil
	}
	return fmt.Errorf("variable not found: %s", strings.Join(fragments, "."))
}

var (
	LowerCaseWithHyphen              = regexp.MustCompilePOSIX("^[a-z][a-z0-9\\-]+$")
	LowerCaseWithUnderscore          = regexp.MustCompilePOSIX("^[a-z][a-z0-9_]+$")
	alphaNumeric                     = regexp.MustCompilePOSIX("^[a-z][a-z0-9]+$")
	camelCase                        = regexp.MustCompilePOSIX("^[a-z][a-z0-9]*([A-Z][a-zA-Z0-9]*)*$")
	LowerCaseWithHyphenAndUnderscore = regexp.MustCompilePOSIX("^[a-z][a-z0-9_\\-]*$")
)

type VariableValidator struct {
	Loop *StepLoop
}

type ExpressionContext int

func NewValidator() ObjectValidation {
	root := NewVariableNode()
	root.Nested("workflow", workflowValidation)

	return root
}

func workflowValidation(workflow *VariableNode) {
	workflow.Next("name")
	workflow.Next("namespace")
	workflow.Next("mainEntrypoint")
	workflow.Next("serviceAccountName")
	workflow.Next("uid")
	workflow.Pattern("parameters", LowerCaseWithHyphenAndUnderscore)
	workflow.Next("parameters")
	workflow.Nested("parameters", func(parameter *VariableNode) {
		parameter.Next("json")
	})
	workflow.Nested("outputs", func(parameter *VariableNode) {
		parameter.Pattern("parameters", regexp.MustCompilePOSIX("^[a-zA-Z][a-zA-Z0-9_\\-]*$"))
		parameter.Pattern("artifacts", regexp.MustCompilePOSIX("^[a-zA-Z][a-zA-Z0-9_\\-]*$"))
	})
	workflow.Nested("annotations", func(parameter *VariableNode) {
		parameter.Next("json")
	})
	workflow.Pattern("annotations", LowerCaseWithHyphenAndUnderscore)
	workflow.Nested("labels", func(parameter *VariableNode) {
		parameter.Next("json")
	})
	workflow.Pattern("labels", LowerCaseWithHyphenAndUnderscore)
	workflow.Next("creationTimestamp")
	workflow.Pattern("creationTimestamp", regexp.MustCompilePOSIX("^%[BbmAadHIMSYypZzL]$"))
	workflow.Nested("creationTimestamp", func(parameter *VariableNode) {
		parameter.Next("RFC3339")
	})
	workflow.Next("priority")
	workflow.Next("duration")
	workflow.Next("scheduledTime")
}
