package main

import (
	"fmt"
	"strings"
)

type VariableValidator struct {
	Loop *StepLoop
}

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

func (v *VariableValidator) Validate(text string) error {
	panic("implement me")
}
