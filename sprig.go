package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
)

type SprigFunction struct {
	Name    string `json:"name"`
	Arity   int    `json:"arity"`
	VarArgs bool   `json:"varargs"`
}

type Sprig struct {
	Functions []SprigFunction `json:"functions"`
}

//go:embed data/run-sprig.json
var sprigJson []byte

func NewSprigFunctions() (*Sprig, error) {
	var sprigConf Sprig
	err := json.Unmarshal(sprigJson, &sprigConf)
	if err != nil {
		return nil, err
	}
	return &sprigConf, nil
}

func (sc Sprig) findFunction(name string) (*SprigFunction, bool) {
	for _, function := range sc.Functions {
		if function.Name == name {
			return &function, true
		}
	}
	return nil, false
}

func (sc Sprig) Validate(name, args string) string {
	function, found := sc.findFunction(name)
	if !found {
		return fmt.Sprintf("unknown sprig function: %s", name)
	}
	arity := len(strings.Split(args, ","))
	if arity < function.Arity || (arity > function.Arity && !function.VarArgs) {
		withVarArgs := ""
		if function.VarArgs {
			withVarArgs = "(varArgs)"
		}
		return fmt.Sprintf("sprig function %s expects %d%s args, got %d", name, function.Arity, withVarArgs, arity)
	}
	return ""
}

// Simplified sprig function validation (example subset)
func validateSprigFunction(funcName, args string) string {
	sprigFuncs := map[string]int{
		"coalesce": -1, // Variable args
		"toString": 1,
	}
	expectedArgs, exists := sprigFuncs[funcName]
	if !exists {
		return fmt.Sprintf("unknown sprig function: %s", funcName)
	}

	argCount := len(strings.Split(args, ",")) - len(strings.Split(args, ", ")) // Rough count
	if expectedArgs != -1 && argCount != expectedArgs {
		return fmt.Sprintf("sprig function %s expects %d args, got %d", funcName, expectedArgs, argCount)
	}
	return ""
}
