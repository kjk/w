package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func getToGenerateFilePath() string {
	path := filepath.Join("cmd", "gengo", "togenerate.txt")
	return path
}

func (g *goGenerator) loadSymbolsToGenerate() {
	path := getToGenerateFilePath()
	d, err := ioutil.ReadFile(path)
	must(err)
	lines := strings.Split(string(d), "\n")
	for _, sym := range lines {
		if len(sym) == 0 {
			continue
		}
		if strings.HasPrefix(sym, "#") {
			continue
		}
		if fi := allFunctions[sym]; fi != nil {
			g.addFunction(sym)
			continue
		}
		if ii := allInterfaces[sym]; ii != nil {
			g.addInterface(sym)
			continue
		}
		fmt.Printf("loadSymbols: unknown symbol '%s'\n", sym)
		os.Exit(1)
	}
}

// must index symbols first
func (g *goGenerator) saveSymbols() {
	var symbols []string
	for _, si := range g.sourceFiles {
		for _, fi := range si.functionsToGenerate {
			symbols = append(symbols, fi.Name)
		}
		for _, ii := range si.interfacesToGenerate {
			symbols = append(symbols, ii.Name())
		}
	}

	moduleToSymbols := map[string][]string{}

	for _, sym := range symbols {
		if fil := allFunctions[sym]; fil != nil {
			fi := fil[0]
			fileName := fi.Module.goSourceFileName()
			a := moduleToSymbols[fileName]
			moduleToSymbols[fileName] = append(a, sym)
			continue
		}
		if ii := allInterfaces[sym]; ii != nil {
			fileName := ii.goSourceFileName()
			a := moduleToSymbols[fileName]
			moduleToSymbols[fileName] = append(a, sym)
			continue
		}
		panic("unknown symbol")
	}

	var keys []string
	for k := range moduleToSymbols {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var lines []string
	for _, k := range keys {
		lines = append(lines, "# "+k)
		names := moduleToSymbols[k]
		sort.Strings(names)
		lines = append(lines, names...)
		lines = append(lines, "")
	}
	s := strings.Join(lines, "\n")
	path := getToGenerateFilePath()
	err := ioutil.WriteFile(path, []byte(s), 0644)
	must(err)
}
