package main

import (
	"fmt"
	"io"
	"iter"
	"os"
	"regexp"
	"strings"

	"github.com/expr-lang/expr"
	jsonpath "github.com/oliveagle/jsonpath"
	"gopkg.in/yaml.v3"
)

// Resource represents a parsed Kubernetes resource
type Resource struct {
	APIVersion string            `yaml:"apiVersion"`
	Kind       string            `yaml:"kind"`
	Metadata   map[string]string `yaml:"metadata"`
	Spec       struct {
		Templates []Template `yaml:"templates"`
	} `yaml:"spec"`
}

// Template represents a template in WorkflowTemplate/Workflow
type Template struct {
	Name  string   `yaml:"name"`
	Steps [][]Step `yaml:"steps"`
}

func (t Template) ParallelSteps() iter.Seq[Step] {
	return func(yield func(Step) bool) {
		for _, ps := range t.Steps {
			for _, s := range ps {
				if !yield(s) {
					return
				}
			}
		}
	}
}

// Step represents a step in a template
type Step struct {
	Name      string      `yaml:"name"`
	Arguments Arguments   `yaml:"arguments"`
	Template  string      `yaml:"template"`
	WithItems interface{} `yaml:"withItems,omitempty"`
}

// Arguments represents arguments passed to a step
type Arguments struct {
	Parameters []Parameter `yaml:"parameters"`
}

// Parameter represents a parameter with name and value
type Parameter struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

// ValidationError represents a linting error
type ValidationError struct {
	FileName     string
	ResourceKind string
	ResourceName string
	TemplateName string
	StepName     string
	ParamName    string
	Value        string
	ErrorMsg     string
}

func main() {
	hasErrors := false
	if len(os.Args) < 2 {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			//goland:noinspection GoUnhandledErrorResult
			fmt.Fprintf(os.Stderr, "error reading stdin: %v\n", err)
			hasErrors = true
		} else {
			err = processData("<stdin>", data)
			if err != nil {
				//goland:noinspection GoUnhandledErrorResult
				fmt.Fprintf(os.Stderr, "error processing stdin: %v\n", err)
				hasErrors = true
			}
		}
	} else {
		files := os.Args[1:]
		for _, file := range files {
			err := processFile(file)
			if err != nil {
				//goland:noinspection GoUnhandledErrorResult
				fmt.Fprintf(os.Stderr, "%s: error: %v\n", file, err)
				hasErrors = true
				continue
			}
		}
	}

	if hasErrors {
		os.Exit(1)
	}
}

func processFile(file string) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	return processData(file, data)
}

func processData(file string, data []byte) error {
	var resource Resource
	err := yaml.Unmarshal(data, &resource)
	if err != nil {
		return fmt.Errorf("failed to parse YAML: %v", err)
	}

	// Check if it's a WorkflowTemplate or Workflow
	if resource.Kind != "WorkflowTemplate" && resource.Kind != "Workflow" {
		fmt.Printf("%s: %s: ok\n", file, resource.Metadata["name"])
		return nil
	}

	errors := lintResource(file, resource)
	if len(errors) == 0 {
		fmt.Printf("%s: %s: ok\n", file, resource.Metadata["name"])
		return nil
	}

	for _, e := range errors {
		fmt.Printf("%s: %s: template=%s, step=%s, param=%s, value=%q: %s\n",
			e.FileName, e.ResourceName, e.TemplateName, e.StepName, e.ParamName, e.Value, e.ErrorMsg)
	}
	return fmt.Errorf("linting errors found")
}

func lintResource(file string, resource Resource) []ValidationError {
	var errors []ValidationError

	for _, template := range resource.Spec.Templates {
		for step := range template.ParallelSteps() {
			for _, param := range step.Arguments.Parameters {
				errMsg := validateValue(param.Value)
				if errMsg != "" {
					errors = append(errors, ValidationError{
						FileName:     file,
						ResourceKind: resource.Kind,
						ResourceName: resource.Metadata["name"],
						TemplateName: template.Name,
						StepName:     step.Name,
						ParamName:    param.Name,
						Value:        param.Value,
						ErrorMsg:     errMsg,
					})
				}
			}
		}
	}
	return errors
}

func validateValue(value string) string {
	// 1. Normal string (no special syntax)
	if !strings.Contains(value, "{{") {
		return ""
	}

	// 2. Workflow Variables ({{...}})
	variableRe := regexp.MustCompile(`^{{[^=].*}}$`)
	if variableRe.MatchString(value) {
		// Basic check; full resolution requires runtime context, so we assume syntax is valid
		return ""
	}

	// 3. jsonpath ({{=jsonpath(...}})
	jsonpathRe := regexp.MustCompile(`^{{=jsonpath\(([^,]+),\s*'([^']+)'\)}}$`)
	if jsonpathRe.MatchString(value) {
		matches := jsonpathRe.FindStringSubmatch(value)
		if len(matches) != 3 {
			return "invalid jsonpath syntax"
		}
		_, err := jsonpath.Compile(matches[2])
		if err != nil {
			return fmt.Sprintf("invalid jsonpath query: %v", err)
		}
		return ""
	}

	// 4. sprig functions ({{=sprig.xxx(...}})
	sprigRe := regexp.MustCompile(`^{{=sprig\.([a-zA-Z]+)\((.*)\)}}$`)
	if sprigRe.MatchString(value) {
		matches := sprigRe.FindStringSubmatch(value)
		if len(matches) != 3 {
			return "invalid sprig function syntax"
		}
		funcName := matches[1]
		args := matches[2]
		return validateSprigFunction(funcName, args)
	}

	// 5. expr-lang ({{=...}})
	exprRe := regexp.MustCompile(`^{{=(.+)}}$`)
	if exprRe.MatchString(value) {
		matches := exprRe.FindStringSubmatch(value)
		if len(matches) != 2 {
			return "invalid expr-lang syntax"
		}
		_, err := expr.Compile(matches[1])
		if err != nil {
			return fmt.Sprintf("invalid expr-lang expression: %v", err)
		}
		return ""
	}

	return "unrecognized value syntax"
}
