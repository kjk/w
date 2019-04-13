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

	WasAdded     bool
	WasGenerated bool
}

// VariableInfo describes a variable
type VariableInfo struct {
	SourceFile *APIMonitorXMLFile
	Headers    *Headers
	Module     *Module    // can be nil
	Condition  *Condition // can be nil
	Variable   *Variable

	WasAdded     bool
	WasGenerated bool
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

	// we keep track of wihch const values have already been generated
	generatedConsts map[string]struct{}

	// the file we're currently writing to
	w io.Writer
}

func newGoGenerator() *goGenerator {
	return &goGenerator{
		modules:         map[string]*goModuleInfo{},
		generatedConsts: map[string]struct{}{},
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

func (g *goGenerator) ws(format string, args ...interface{}) {
	s := format
	if len(args) > 0 {
		s = fmt.Sprintf(format, args...)
	}
	ws(g.w, s)
}

func (g *goGenerator) generateAlias(vi *VariableInfo) {
	if vi.WasGenerated {
		return
	}

	v := vi.Variable
	if v.Name[0] == '[' {
		return
	}

	g.ws("type %s %s\n", v.Name, v.Base)
	vi.WasGenerated = true
}

func (g *goGenerator) generateConsts(set []*Set) {
	// TODO: optimize if only one value
	if len(set) == 0 {
		return
	}
	g.ws("const (\n")
	for _, e := range set {
		name := e.Name
		// the same value can appear in multiple enum sets
		// so we need to ensure we only generate value once
		if _, ok := g.generatedConsts[name]; ok {
			continue
		}
		g.ws("%s = %s\n", name, e.Value)
		g.generatedConsts[name] = struct{}{}
	}
	g.ws(")\n\n")
}

func (g *goGenerator) generateSet(vi *VariableInfo) {
	// TODO: can have Flag and Enum and Set in same Variable?
	if vi.WasGenerated {
		return
	}

	v := vi.Variable
	vi.WasGenerated = true
	g.generateConsts(v.Set)
}

func (g *goGenerator) generateEnum(vi *VariableInfo) {
	// TODO: can have Flag and Enum and Set in same Variable?
	if vi.WasGenerated {
		return
	}

	v := vi.Variable
	if v.Enum == nil {
		return
	}

	vi.WasGenerated = true
	g.generateConsts(v.Enum.Set)
}

func (g *goGenerator) generateFlag(vi *VariableInfo) {
	// TODO: can have Flag and Enum and Set in same Variable?
	if vi.WasGenerated {
		return
	}

	v := vi.Variable
	if v.Flag == nil {
		return
	}

	vi.WasGenerated = true
	g.generateConsts(v.Flag.Set)
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

	for _, fi := range mi.functions {
		fnName := getFunctionName(fi.Function)
		varName := funcNameToVarName(fnName)
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
		fnName := getFunctionName(fi.Function)
		varName := funcNameToVarName(fnName)
		g.ws("\t%s = lib%s.NewProc(\"%s\")\n", varName, dllNameNoExt, fnName)
	}

	g.ws("}\n")

	for _, fi := range mi.functions {
		g.genFunction(fi)
	}
}

func isWide(fn *Api) bool {
	// when we have both W and A versions, do W
	if fn.BothCharset == "True" {
		return true
	}
	// TODO: more cases ?
	return false
}

func getFunctionName(fn *Api) string {
	if isWide(fn) {
		return fn.Name + "W"
	}
	return fn.Name
}

func (g *goGenerator) desugarTypeNamed(tp string) string {
	if predefined := desugerPreDefinedType(tp); predefined != "" {
		return predefined
	}

	vi := findVariable(tp)
	if vi == nil {
		s := fmt.Sprintf("didn't find info about type '%s'\n", tp)
		panic(s)
	}
	return g.desugarType(vi)
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
	case "int":
		// yes, it's strange but that's what it seems to be
		return "int32"
	case "ModuleHandle":
		return "HANDLE"
	}
	return ""
}

// this converts type to a real Go type we want to use
func (g *goGenerator) desugarType(vi *VariableInfo) string {
	v := vi.Variable
	tp := v.Type

	if predefined := desugerPreDefinedType(tp); predefined != "" {
		return predefined
	}

	switch tp {
	case "Integer":
		return desugarInteger(v)
	}

	g.generateFlag(vi)
	g.generateSet(vi)
	g.generateEnum(vi)

	if tp == typeAlias {
		g.generateAlias(vi)
		// we want to preserve types that are aliases for HANDLE
		// (HWND, HMENU etc.)
		if v.Base == "HANDLE" {
			return v.Name
		}
		return g.desugarTypeNamed(v.Base)
	}

	if tp == typePointer {
		return "*" + g.desugarTypeNamed(v.Base)
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

	// first a pass to generate types this function depends on
	for _, arg := range fn.Params {
		g.desugarTypeNamed(arg.Type)
	}
	returnType := ""
	hasReturn := fn.Return != nil
	if hasReturn {
		// TODO: to handle BOOL => bool need a version of desugarTypeNamed()
		// specialized for return types
		returnType = g.desugarTypeNamed(fn.Return.Type)
	}

	fnName := getFunctionName(fn)
	g.ws("\nfunc %s(", fnName)
	lastIdx := len(fn.Params) - 1
	for idx, arg := range fn.Params {
		name := arg.Name
		g.ws(name + " ")
		typ := g.desugarTypeNamed(arg.Type)
		g.ws(typ)
		if idx != lastIdx {
			g.ws(", ")
		}
	}

	g.ws(")")
	if hasReturn {
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
	g.ws(" := %s(%s.Addr(), %d,\n", sysName, fnVarName, nArgs)
	for _, arg := range fn.Params {
		if g.isPointerType(arg.Type) {
			g.ws("uintptr(unsafe.Pointer(%s)),\n", arg.Name)
		} else {
			g.ws("uintptr(%s),\n", arg.Name)
		}
	}
	nLeftOver := sysArgsCount - nArgs
	for nLeftOver > 0 {
		g.ws("0,\n")
	}
	g.ws(")\n")
	if hasReturn {
		g.ws("return %s(ret)\n", returnType)
	}

	g.ws("\n}")
}

func (g *goGenerator) isPointerType(tp string) bool {
	typ := g.desugarTypeNamed(tp)
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
