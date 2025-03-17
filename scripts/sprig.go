package main

import (
	"encoding/json"
	"fmt"
	"github.com/Masterminds/sprig/v3"
	"os"
	"reflect"
)

type Fun struct {
	Name    string `json:"name"`
	Arity   int    `json:"arity"`
	VarArgs bool   `json:"varargs"`
}

type FunCollection struct {
	Functions []Fun `json:"functions"`
}

func main() {
	functions := sprig.GenericFuncMap()
	funs := make([]Fun, 0)
	for name, function := range functions {
		mayFunction := reflect.TypeOf(function)
		if mayFunction.Kind() == reflect.Func {
			fun := mayFunction
			arity := fun.NumIn()
			varArgs := false
			if 0 < arity {
				lastParam := fun.In(arity - 1)
				lastParamKind := lastParam.Kind()
				varArgs = lastParamKind == reflect.Slice
			}
			funs = append(funs, Fun{
				Name:    name,
				Arity:   arity,
				VarArgs: varArgs,
			})
		}
	}
	fc := FunCollection{funs}
	bs, err := json.Marshal(&fc)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "failed to write json", err)
		os.Exit(1)
	}
	fmt.Println(string(bs))
}
