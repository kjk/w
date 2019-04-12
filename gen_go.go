package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	// names for Variable.Type field
	typeAlias     = "Alias"
	typePtr       = "Pointer"
	typePointer   = "Pointer"
	typeInterface = "Interface"
)

// FunctionInfo describes a function
type FunctionInfo struct {
	SourceFile *APIMonitorXMLFile
	Module     *Module
	Function   *Api

	WasAdded bool
}

// VariableInfo describes a variable
type VariableInfo struct {
	SourceFile *APIMonitorXMLFile
	Headers    *Headers
	Module     *Module    // can be nil
	Condition  *Condition // can be nil
	Variable   *Variable

	WasAdded bool
}

var (
	functions = map[string][]*FunctionInfo{}
	variables = map[string][]*VariableInfo{}
)

//var toGen = []string{"CreateWindowEx"}

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

func findFunction(name string) *FunctionInfo {
	a := functions[name]
	if len(a) == 0 {
		return nil
	}
	if len(a) > 1 {
		fmt.Printf("Found %d functions with name '%s'\n", len(a), name)
	}
	return a[0]
}

func findVariable(name string) *VariableInfo {
	a := variables[name]
	if len(a) == 0 {
		return nil
	}
	if len(a) > 1 {
		fmt.Printf("Found %d variables with name '%s', showing the first one\n", len(a), name)
	}
	return a[0]
}

func gofmtFile(path string) {
	cmd := exec.Command("gofmt", "-s", "-w", path)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("gofmt failed with %s. Output:\n%s\n", err, string(out))
	}
}

const (
	fileHdr = `package w

import (
	"golang.org/x/sys/windows"
	"syscall"
	"unsafe"
)
`
)

var (
	// this is for Variable that don't belong to any module
	noModule = "nomodule"
)

// information needed to generate info for a single module
// (aka dll).
type goModuleInfo struct {
	name      string // e.g. gdi32
	functions []*FunctionInfo
	variables []*VariableInfo

	generatedFilePath string
}

// goGenerator keeps info needed for generating Go files
// we generate one .go file per module
type goGenerator struct {
	// maps module (dll) name to info for generating .go
	// file for that module
	modules map[string]*goModuleInfo

	// the file we're currently writing to
	w io.Writer
}

func newGoGenerator() *goGenerator {
	return &goGenerator{
		modules: map[string]*goModuleInfo{},
	}
}

func (g *goGenerator) addVariable(vi *VariableInfo) {
	if vi.WasAdded {
		return
	}
	v := vi.Variable
	//serVariable(v, 0)
	vi.WasAdded = true
	tp := strings.ToLower(v.Type)
	switch tp {
	case "pointer", "alias", "enum", "flag", "array":
		// fmt.Printf("Type: %s, base: %s\n", v.Type, v.Base)
		g.addSymbol(v.Base)
	case "integer", "modulehandle", "void", "tcharacter":
		// skip those
	default:
		fmt.Printf("Type: %s\n", v.Type)
	}
}

func (g *goGenerator) getModuleInfo(mod *Module) *goModuleInfo {
	name := noModule
	if mod != nil {
		name = mod.Name
	}
	mi := g.modules[name]
	if mi == nil {
		mi = &goModuleInfo{
			name: name,
		}
		g.modules[name] = mi
	}
	return mi
}

func (g *goGenerator) addFunction(fi *FunctionInfo) {
	if fi.WasAdded {
		return
	}
	mi := g.getModuleInfo(fi.Module)
	mi.functions = append(mi.functions, fi)
	fi.WasAdded = true
	//serApi(fi.Function, 0)
	ret := fi.Function.Return
	g.addSymbol(ret.Type)
	for _, arg := range fi.Function.Params {
		g.addSymbol(arg.Type)
	}
}

func (g *goGenerator) addSymbol(name string) {
	// predefined types are alredy in
	if predefined := desugerPreDefinedType(name); predefined != "" {
		return
	}

	fi := findFunction(name)
	if fi != nil {
		g.addFunction(fi)
		return
	}
	vi := findVariable(name)
	if vi != nil {
		g.addVariable(vi)
		return
	}
	s := fmt.Sprintf("Didn't find function or variable with name '%s'", name)
	panic(s)
}

func ws(w io.Writer, s string) {
	_, err := io.WriteString(w, s)
	must(err)
}

func (g *goGenerator) ws(s string) {
	ws(g.w, s)
}

func (g *goGenerator) generateModule(mi *goModuleInfo) {
	fmt.Printf("Module: %s\n", mi.name)
	dllName := strings.ToLower(mi.name)
	dllNameNoExt := strings.TrimSuffix(dllName, ".dll")
	fileName := dllNameNoExt + ".go"
	path := filepath.Join("generated", fileName)
	mi.generatedFilePath = path
	var err error
	g.w, err = os.Create(path)
	must(err)
	defer func() {
		f := g.w.(*os.File)
		err := f.Close()
		must(err)
		g.w = nil
	}()
	g.ws(fileHdr)

	s := fmt.Sprintf(`
var (
	lib%s *windows.LazyDLL
)
`, dllNameNoExt)
	g.ws(s)

	/*
	   var (
	   	abortDoc               *windows.LazyProc
	   	addFontResourceEx      *windows.LazyProc
	   	alphaBlend             *windows.LazyProc
	   )
	*/
	g.ws("\nvar (\n")

	for _, fn := range mi.functions {
		varName := funcNameToVarName(fn.Function.Name)
		s = fmt.Sprintf("\t%s *windows.LazyProc\n", varName)
		g.ws(s)
	}

	g.ws("\n)\n")
	/*
		func init() {
			// Library
			libkernel32 = windows.NewLazySystemDLL("kernel32.dll")

			// Functions
			activateActCtx = libkernel32.NewProc("ActivateActCtx")
		}
	*/
	g.ws("\nfunc init() {\n")
	s = fmt.Sprintf(`
	lib%s = windows.NewLazySystemDLL("%s.dll")

`, dllNameNoExt, dllNameNoExt)
	g.ws(s)

	for _, fi := range mi.functions {
		varName := funcNameToVarName(fi.Function.Name)
		fnName := getFunctionName(fi.Function)
		s = fmt.Sprintf("\t%s = lib%s.NewProc(\"%s\")\n", varName, dllNameNoExt, fnName)
		g.ws(s)
	}

	g.ws("}\n")

	for _, fi := range mi.functions {
		g.genFunction(fi)
	}
}

func getFunctionName(fn *Api) string {
	name := fn.Name
	// TODO: not sure this condition is good enough
	if fn.BothCharset != "" { // it's True
		// fmt.Printf("BothCharset: '%s'\n", fn.Function.BothCharset)
		name += "W"
	}
	return name
}

func desugarTypeNamed(tp string) string {
	if predefined := desugerPreDefinedType(tp); predefined != "" {
		return predefined
	}

	vi := findVariable(tp)
	if vi == nil {
		s := fmt.Sprintf("didn't find info about type '%s'\n", tp)
		panic(s)
	}
	return desugarType(vi)
}

func desugarInteger(v *Variable) string {
	tp := "int"
	if v.Unsigned == "True" {
		tp = "u" + tp
	}
	switch v.Size {
	case "1":
		return tp + "8"
	case "2":
		return tp + "16"
	case "4":
		return tp + "32"
	case "8":
		return tp + "64"
	}

	//pretty.Print(v)
	panic("unsupported Variable of type Integer")
}

// if returns "", not a known type
func desugerPreDefinedType(tp string) string {
	// some known terminal types
	switch tp {
	case "LPVOID":
		return "unsafe.Pointer"
	case "LPCTSTR":
		return "*uint16"
	case "int8", "int16", "int32", "int64":
		return tp
	case "uint8", "uint16", "uint32", "uint64":
		return tp
	case "UINT_PTR":
		return "uintptr"
	case "ModuleHandle":
		return "HANDLE"
	}
	return ""
}

// this converts type to a real Go type we want to use
func desugarType(vi *VariableInfo) string {
	v := vi.Variable
	tp := v.Type

	if predefined := desugerPreDefinedType(tp); predefined != "" {
		return predefined
	}

	switch tp {
	case "Integer":
		return desugarInteger(v)
	}

	if tp == typeAlias {
		// we want to preserve types that are aliases for HANDLE
		// (HWND, HMENU etc.)
		if v.Base == "HANDLE" {
			return v.Name
		}
		return desugarTypeNamed(v.Base)
	}

	if tp == typePointer {
		return "*" + desugarTypeNamed(v.Base)
	}

	// TODO: recursively resolve the type
	return tp
}

/*
func CreateWindowEx(dwExStyle uint32, lpClassName, lpWindowName *uint16, dwStyle uint32, x, y, nWidth, nHeight int32, hWndParent HWND, hMenu HMENU, hInstance HINSTANCE, lpParam unsafe.Pointer) HWND {
	ret, _, _ := syscall.Syscall12(createWindowEx.Addr(), 12,
		uintptr(dwExStyle),
		uintptr(unsafe.Pointer(lpClassName)),
		uintptr(unsafe.Pointer(lpWindowName)),
		uintptr(dwStyle),
		uintptr(x),
		uintptr(y),
		uintptr(nWidth),
		uintptr(nHeight),
		uintptr(hWndParent),
		uintptr(hMenu),
		uintptr(hInstance),
		uintptr(lpParam))

	return HWND(ret)
}
*/
func (g *goGenerator) genFunction(fi *FunctionInfo) {
	fn := fi.Function
	fnName := getFunctionName(fn)
	g.ws("\nfunc " + fnName + "(")
	lastIdx := len(fn.Params) - 1
	for idx, arg := range fn.Params {
		name := arg.Name
		g.ws(name + " ")
		typ := desugarTypeNamed(arg.Type)
		g.ws(typ)
		if idx != lastIdx {
			g.ws(", ")
		}
	}

	g.ws(")")
	returnType := ""
	hasReturn := fn.Return != nil
	if hasReturn {
		// TODO: to handle BOOL => bool need a version of desugarTypeNamed()
		// specialized for return types
		returnType = desugarTypeNamed(fn.Return.Type)
		g.ws(" " + returnType)
	}
	g.ws(" {\n")

	// 	ret, _, _ := syscall.Syscall12(createWindowEx.Addr(), 12,
	fnVarName := funcNameToVarName(fnName)
	nArgs := len(fn.Params)
	sysName, sysArgsCount := syscallFuncForArgCount(nArgs)
	if hasReturn {
		g.ws("ret, _, _")
	} else {
		g.ws("_, _, _")
	}
	g.ws(fmt.Sprintf(" := %s(%s.Addr(), %d,\n", sysName, fnVarName, nArgs))
	for _, arg := range fn.Params {
		if isPointerType(arg.Type) {
			g.ws(fmt.Sprintf("uintptr(unsafe.Pointer(%s)),\n", arg.Name))
		} else {
			g.ws(fmt.Sprintf("uintptr(%s),\n", arg.Name))
		}
	}
	nLeftOver := sysArgsCount - nArgs
	for nLeftOver > 0 {
		g.ws("0,\n")
	}
	g.ws(")\n")
	if hasReturn {
		g.ws(fmt.Sprintf("return %s(ret)\n", returnType))
	}

	g.ws("\n}")
}

func isPointerType(tp string) bool {
	typ := desugarTypeNamed(tp)
	if typ[0] == '*' {
		return true
	}
	return false
}

func (g *goGenerator) generate() {
	var modules []string
	for mod := range g.modules {
		modules = append(modules, mod)
	}
	sort.Strings(modules)
	for _, name := range modules {
		mi := g.modules[name]
		g.generateModule(mi)
		gofmtFile(mi.generatedFilePath)
		dumpFile(mi.generatedFilePath)
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

	g := newGoGenerator()
	g.addSymbol("CreateWindowEx")
	g.generate()
}
