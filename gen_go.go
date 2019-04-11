package main

import (
	"fmt"
	"os"
	"time"
)

// FunctionInfo is infor about a function
type FunctionInfo struct {
	Function   *Api
	Module     *Module
	SourceFile *APIMonitorXMLFile
}

var (
	functions = map[string]*FunctionInfo{}
)

var toGen = []string{"CreateWindowEx"}

func shortApiName(fn *FunctionInfo) string {
	sfn := fn.SourceFile.FileName
	modName := fn.Module.Name
	fnName := fn.Function.Name
	return fmt.Sprintf(`%s %s.%s`, sfn, modName, fnName)
}

func indexFunction(api *Api, mod *Module, f *APIMonitorXMLFile) {
	name := api.Name
	fn := &FunctionInfo{
		Function:   api,
		Module:     mod,
		SourceFile: f,
	}
	if curr, ok := functions[name]; ok {
		// Possible reasons for duplication:
		// - 2 entries for A and W versions (e.g. InternetGetConnectedStateEx)
		// - function in different library on different OSes
		//   e.g. IcmpCreateFile
		if false {
			currName := shortApiName(curr)
			newName := shortApiName(fn)
			fmt.Printf("Found duplicate function:\n  %s\n  %s\n", currName, newName)
		}
		return
	}
	functions[name] = fn
}

func buildIndex(files []*APIMonitorXMLFile) {
	for _, f := range files {
		if shouldSkipFile(f.FileName) {
			continue
		}
		for _, mod := range f.Modules {
			for _, api := range mod.Apis {
				indexFunction(api, mod, f)
			}
		}
	}
}

func findSymbol(xmlFiles []*APIMonitorXMLFile, name string) {
	fn := functions[name]
	if fn == nil {
		fmt.Printf("Dind't find symbol '%s'\n", name)
		return
	}
	s := shortApiName(fn)
	fmt.Printf("Found function '%s'\n", s)
	serApi(fn.Function, 0)
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
	fmt.Printf("Built index in %s\n", time.Since(timeStart))

	findSymbol(parsedFiles, "CreateWindowEx")
}
