package main

import (
	"fmt"
	"os"
	"time"
)

// FunctionInfo describes a function
type FunctionInfo struct {
	SourceFile *APIMonitorXMLFile
	Module     *Module
	Function   *Api
}

// VariableInfo describes a variable
type VariableInfo struct {
	SourceFile *APIMonitorXMLFile
	Headers    *Headers
	Module     *Module    // can be nil
	Condition  *Condition // can be nil
	Variable   *Variable
}

var (
	functions = map[string][]*FunctionInfo{}
	variables = map[string][]*VariableInfo{}
)

var toGen = []string{"CreateWindowEx"}

func shortAPIName(fn *FunctionInfo) string {
	sfn := fn.SourceFile.FileName
	modName := fn.Module.Name
	fnName := fn.Function.Name
	return fmt.Sprintf(`%s %s.%s`, sfn, modName, fnName)
}

func indexFunction(f *APIMonitorXMLFile, mod *Module, api *Api) {
	name := api.Name
	fi := &FunctionInfo{
		Function:   api,
		Module:     mod,
		SourceFile: f,
	}
	a := functions[name]
	functions[name] = append(a, fi)
	/*
		if curr, ok := functions[name]; ok {
			// Possible reasons for duplication:
			// - 2 entries for A and W versions (e.g. InternetGetConnectedStateEx)
			// - function in different library on different OSes
			//   e.g. IcmpCreateFile
			if false {
				currName := shortAPIName(curr)
				newName := shortAPIName(fn)
				fmt.Printf("Found duplicate function:\n  %s\n  %s\n", currName, newName)
			}
			return
		}
	*/
}

func indexVariable(f *APIMonitorXMLFile, hdrs *Headers, mod *Module, cond *Condition, v *Variable) {
	{
		name := v.Name
		vi := &VariableInfo{
			SourceFile: f,
			Headers:    hdrs,
			Module:     mod,
			Condition:  cond,
			Variable:   v,
		}
		a := variables[name]
		variables[name] = append(a, vi)
	}

	/*
		if v.Enum != nil {
			name := v.Enum.DefaultName
			vi := &VariableInfo{
				SourceFile: f,
				Headers:    hdrs,
				Variable:   v,
			}
			a := variables[name]
			variables[name] = append(a, vi)
		}
	*/
}

func indexModules(f *APIMonitorXMLFile) {
	if f.Modules == nil {
		return
	}
	for _, mod := range f.Modules {
		for _, api := range mod.Apis {
			indexFunction(f, mod, api)
		}
		for _, v := range mod.Variables {
			indexVariable(f, f.Headers, mod, nil, v)
		}
	}
}

func indexHeaders(f *APIMonitorXMLFile) {
	if f.Headers == nil {
		return
	}
	for _, cond := range f.Headers.Conditions {
		for _, v := range cond.Variable {
			indexVariable(f, f.Headers, nil, cond, v)
		}
	}
	for _, v := range f.Headers.Variables {
		indexVariable(f, f.Headers, nil, nil, v)
	}
}

func buildIndex(files []*APIMonitorXMLFile) {
	for _, f := range files {
		if shouldSkipFile(f.FileName) {
			continue
		}
		indexModules(f)
		indexHeaders(f)
	}
}

func findFunction(name string) bool {
	a := functions[name]
	if len(a) == 0 {
		return false
	}
	fn := a[0]
	s := shortAPIName(fn)
	fmt.Printf("Found function '%s'\n", s)
	serApi(fn.Function, 0)
	return true
}

func findVariable(name string) bool {
	a := variables[name]
	if len(a) == 0 {
		return false
	}
	if len(a) > 0 {
		fmt.Printf("Found %d variables with name '%s', showing the first one\n", len(a), name)
	}
	vi := a[0]
	fmt.Printf("Found variable '%s'\n", name)
	serVariable(vi.Variable, 0)
	return true
}

func findSymbol(name string) {
	found1 := findFunction(name)
	found2 := findVariable(name)
	ok := found1 || found2
	if !ok {
		fmt.Printf("Dind't find symbol '%s'\n", name)
	}
}

func genGo() {
	fmt.Printf("Starting gen go\n")
	out = os.Stdout

	timeStart := time.Now()
	parsedFiles, err := parseApiMonitorData()
	must(err)
	fmt.Printf("Parsed %d files in %s\n", len(parsedFiles), time.Since(timeStart))

	timeStart = time.Now()
	buildIndex(parsedFiles)
	fmt.Printf("Built index in %s. %d functions, %d variables\n", time.Since(timeStart), len(functions), len(variables))

	findSymbol("CreateWindowEx")
	findSymbol("DLGTEMPLATE")
	findSymbol("[WindowExStyle]")
}
