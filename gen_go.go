package main

import (
	"fmt"
	"time"

	"github.com/kylelemons/godebug/pretty"
)

type FunctionInfo struct {
	Module   *Module
	Function *Api
}

var ()
var toGen = []string{"CreateWindowEx"}

func findSymbolInModule(mod *Module, name string) bool {
	for _, api := range mod.Apis {
		if api.Name == name {
			fmt.Printf("Found function '%s' in module %s\n", name, mod.Name)
			pretty.Print(api)
			return true
		}
	}
	return false
}

func findSymbolInFile(xmlFile *ApiMonitorXmlFile, name string) bool {
	data := xmlFile
	for _, mod := range data.Modules {
		found := findSymbolInModule(mod, name)
		if found {
			return true
		}
	}
	return false
}

func findSymbol(xmlFiles []*ApiMonitorXmlFile, name string) {
	for _, f := range xmlFiles {
		found := findSymbolInFile(f, name)
		if found {
			fmt.Printf("Found\n")
			return
		}
	}
}

func genGo() {
	fmt.Printf("Starting gen go\n")
	timeStart := time.Now()
	parsedFiles, err := parseApiMonitorData()
	must(err)
	fmt.Printf("Parsed %d files in %s\n", len(parsedFiles), time.Since(timeStart))
	findSymbol(parsedFiles, "CreateWindowEx")
}
