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
	debugDescription(ident int) string
}

type VariableNode struct {
	Names     []string
	Children  map[string]ObjectValidation
	Patterned []*PatternedVariableNameNodeValidator
}

func (v *VariableNode) debugDescription(ident int) string {
	var sb strings.Builder
	indent := strings.Repeat("    ", ident)
	for _, name := range v.Names {
		sb.WriteString(indent)
		sb.WriteString("  n: ")
		sb.WriteString(name)
		sb.WriteRune('\n')
	}
	for name, child := range v.Children {
		sb.WriteString(indent)
		sb.WriteString("  c: ")
		sb.WriteString(name)
		sb.WriteRune('\n')
		sb.WriteString(child.debugDescription(ident + 1))
	}
	sb.WriteString(indent)
	return sb.String()
}

func NewVariableNode() *VariableNode {
	names := make([]string, 0)
	children := make(map[string]ObjectValidation)
	patterned := make([]*PatternedVariableNameNodeValidator, 0)
	return &VariableNode{
		Names:     names,
		Children:  children,
		Patterned: patterned,
	}
}

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
	for _, validation := range v.Patterned {
		if validation.Pattern.Match([]byte(fragment)) {
			return validation.Validate(index, fragments)
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
	} else if composite, ok := ov.(*CompositeValidator); ok {
		child := NewVariableNode()
		config(child)
		composite.validators = append(composite.validators, child)
	} else {
		child := NewVariableNode()
		config(child)
		validators := []ObjectValidation{
			ov,
			child,
		}
		v.Children[name] = &CompositeValidator{validators}
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
	if ov, ok := v.Children[name]; !ok {
		v.Children[name] = child
	} else if composite, ok := ov.(*CompositeValidator); ok {
		composite.validators = append(composite.validators, child)
	} else {
		validators := []ObjectValidation{
			ov,
			child,
		}
		v.Children[name] = &CompositeValidator{validators}
	}
}

func (v *VariableNode) PatternNested(npattern *regexp.Regexp, config func(variableNode *VariableNode)) {
	node := NewVariableNode()
	config(node)
	child := &PatternedVariableNameNodeValidator{
		Pattern: npattern,
		node:    node,
	}
	v.Patterned = append(v.Patterned, child)
}

type PatternedVariableNameNodeValidator struct {
	Pattern *regexp.Regexp
	node    ObjectValidation
}

func (p *PatternedVariableNameNodeValidator) debugDescription(ident int) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%sc: %s\n", strings.Repeat("    ", ident), p.Pattern.String()))
	sb.WriteString(p.debugDescription(ident + 1))
	return sb.String()
}

func (p *PatternedVariableNameNodeValidator) Validate(index int, fragments []string) error {
	m := len(fragments)
	if m <= index {
		return fmt.Errorf("invalid index: %d", index)
	}
	if index == m {
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

func (p *PatternedVariableNameLeafValidator) debugDescription(ident int) string {
	return fmt.Sprintf("%sp: %s\n", strings.Repeat("    ", ident), p.Pattern.String())
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

type CompositeValidator struct {
	validators []ObjectValidation
}

func (v *CompositeValidator) debugDescription(ident int) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%scomposite:\n", strings.Repeat("    ", ident)))
	for _, validator := range v.validators {
		sb.WriteString(validator.debugDescription(ident + 1))
	}
	return sb.String()
}

func (v *CompositeValidator) Validate(index int, fragments []string) error {
	m := len(fragments)
	if index == m {
		return fmt.Errorf("variable not found: %s", strings.Join(fragments, "."))
	}
	for _, validator := range v.validators {
		err := validator.Validate(index, fragments)
		if err == nil {
			return nil
		}
	}
	return fmt.Errorf("variable not found: %s", strings.Join(fragments, "."))
}

var (
	LowerCaseWithHyphen              = regexp.MustCompilePOSIX("^[a-z][a-z0-9\\-]+$")
	LowerCaseWithUnderscore          = regexp.MustCompilePOSIX("^[a-z][a-z0-9_]+$")
	alphaNumeric                     = regexp.MustCompilePOSIX("^[a-z][a-z0-9]+$")
	camelCase                        = regexp.MustCompilePOSIX("^[a-z][a-z0-9]*([A-Z][a-zA-Z0-9]*)*$")
	LowerCaseWithHyphenAndUnderscore = regexp.MustCompilePOSIX("^[a-z][a-z0-9_\\-]*$")
	BothCasesWithHyphen              = regexp.MustCompilePOSIX("^[a-zA-Z][a-zA-Z0-9]*(-[a-zA-Z][a-zA-Z0-9]*)*$")
)

type VariableValidator struct {
	Loop *StepLoop
}

type ExpressionContext int

const (
	AllTemplates ExpressionContext = iota
	GlobalExpression
	StepsTemplates
	DAGTemplates
	HTTPTemplates
	CronWorkflows
	RetryStrategies
	ContainerScriptTemplates
	LoopsTemplates
	MetricsTemplates
	WorkflowMetrics
	TemplateMetrics
)

func (e ExpressionContext) ConfigureValidation(root *VariableNode) {
	switch e {
	case AllTemplates:
		root.Nested("inputs", inputsValidation)
		root.Nested("node", func(node *VariableNode) {
			node.Next("name")
		})
		break
	case GlobalExpression:
		root.Nested("workflow", workflowValidation)
		break
	case StepsTemplates:
		root.Nested("steps", stepsValidation)
		break
	case DAGTemplates:
		root.Nested("tasks", tasksValidation)
		break
	case HTTPTemplates:
		root.Nested("request", httpRequestValidation)
		root.Nested("response", httpResponseValidation)
	default:
	}
}

func NewValidator(expressionContexts ...ExpressionContext) ObjectValidation {
	root := NewVariableNode()
	for _, ec := range expressionContexts {
		ec.ConfigureValidation(root)
	}
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

func inputsValidation(inputs *VariableNode) {
	inputs.Next("parameters")
	inputs.Pattern("parameters", LowerCaseWithHyphenAndUnderscore)
	inputs.Pattern("artifacts", LowerCaseWithHyphenAndUnderscore)
}

func stepsValidation(steps *VariableNode) {
	steps.Next("name")
	steps.PatternNested(LowerCaseWithHyphen, func(step *VariableNode) {
		step.Next("id")
		step.Next("ip")
		step.Next("status")
		step.Next("exitCode")
		step.Next("startedAt")
		step.Next("finishedAt")
		step.Next("hostNodeName")
		step.Nested("outputs", func(outputs *VariableNode) {
			outputs.Next("parameters")
			outputs.Next("result")
			outputs.Pattern("parameters", LowerCaseWithHyphenAndUnderscore)
			outputs.Pattern("artifacts", LowerCaseWithHyphenAndUnderscore)
		})
	})
}

func tasksValidation(tasks *VariableNode) {
	tasks.Next("name")
	tasks.PatternNested(LowerCaseWithHyphen, func(task *VariableNode) {
		task.Next("id")
		task.Next("ip")
		task.Next("status")
		task.Next("exitCode")
		task.Next("startedAt")
		task.Next("finishedAt")
		task.Next("hostNodeName")
		task.Nested("outputs", func(outputs *VariableNode) {
			outputs.Next("parameters")
			outputs.Next("result")
			outputs.Pattern("parameters", LowerCaseWithHyphenAndUnderscore)
			outputs.Pattern("artifacts", LowerCaseWithHyphenAndUnderscore)
		})
	})
}

func httpRequestValidation(request *VariableNode) {
	request.Next("method")
	request.Next("url")
	request.Next("body")
	request.Next("headers")
	request.Pattern("headers", BothCasesWithHyphen)
}

func httpResponseValidation(response *VariableNode) {
	response.Next("statusCode")
	response.Next("body")
	response.Next("headers")
	response.Pattern("headers", BothCasesWithHyphen)
}
