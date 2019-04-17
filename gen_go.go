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
	typeStruct    = "Struct"
	typeUnion     = "Union"
	typeArray     = "Array"
	typeVoid      = "Void"
)

// TypeInfo describes a type
type TypeInfo struct {
	SourceFile *APIMonitorXMLFile
	Headers    *Headers
	Module     *Module    // can be nil
	Condition  *Condition // can be nil
	Variable   *Variable

	// this is a name of Go type representig this type
	// after desugaring type (e.g. resolving LPWSTR => *uint16)
	TypeName string

	WasAdded     bool
	WasGenerated bool
}

// FunctionArgument describes an argument to a function
type FunctionArgument struct {
	Name string
	Type *TypeInfo
}

// FunctionInfo describes a function
type FunctionInfo struct {
	SourceFile *APIMonitorXMLFile
	Module     *Module
	Function   *Api

	Name string
	// if there is both A and W version of the function,
	// it indicates this is Unicode (W) version
	IsUnicode bool

	Args []*FunctionArgument
	// if nil, no return type i.e. void
	ReturnType *TypeInfo

	WasAdded     bool
	WasGenerated bool
}

// GoVarName returns name of the global variable that represents
// this function e.g. AbortDoc => abortDoc
func (f *FunctionInfo) GoVarName() string {
	c := f.Name[:1]
	c = strings.ToLower(c)
	return c + f.Name[1:]
}

// InterfaceInfo describes an interface
type InterfaceInfo struct {
	// name of the file that refers to this interface
	// e.g. "URL.h.xml"
	FileName  string
	Interface *Interface
	// Interface.Api but with more info
	Methods []*FunctionInfo
	// Interface.Variable but with more info
	Types []*TypeInfo

	WasAdded     bool
	WasGenerated bool
}

func shortAPIName(fn *FunctionInfo) string {
	sfn := fn.SourceFile.FileName
	modName := fn.Module.Name
	fnName := fn.Function.Name
	return fmt.Sprintf(`%s %s.%s`, sfn, modName, fnName)
}

func findFunction(name string) *FunctionInfo {
	a := allFunctions[name]
	if len(a) == 0 {
		return nil
	}
	if len(a) > 1 {
		fmt.Printf("Found %d functions with name '%s'\n", len(a), name)
	}
	return a[0]
}

func findType(name string) *TypeInfo {
	a := allTypes[name]
	if len(a) == 0 {
		return nil
	}
	if len(a) > 1 {
		fmt.Printf("Found %d variables with name '%s', showing the first one\n", len(a), name)
	}
	return a[0]
}

func findInterface(name string) *InterfaceInfo {
	return allInterfaces[name]
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
	types     []*TypeInfo

	generatedFilePath string
}

// goGenerator keeps info needed for generating Go files
// we generate one .go file per module
type goGenerator struct {
	// maps module (dll) name to info for generating .go
	// file for that module
	modules map[string]*goModuleInfo

	// list of interfaces to add
	interfaces []string

	// we keep track of which const values have already been generated
	// because they the same constant might belong to multiple enums / sets
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

func (g *goGenerator) addType(vi *TypeInfo) {
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
	case "struct", "interface":
		// also skip those
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

func (g *goGenerator) rememberFunction(fi *FunctionInfo) {
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

func (g *goGenerator) addInterface(ii *InterfaceInfo) {
	if ii.WasAdded {
		return
	}
	name := ii.Interface.Name
	g.interfaces = append(g.interfaces, name)
	ii.WasAdded = true
}

func (g *goGenerator) addFunctionMust(name string) {
	added := g.addFunction(name)
	panicIf(!added, "didn't find function '%s'", name)
}

func (g *goGenerator) addFunction(name string) bool {
	fi := findFunction(name)
	if fi != nil {
		g.rememberFunction(fi)
		return true
	}
	return false
}

func (g *goGenerator) addSymbol(name string) {
	// predefined types are alredy in
	if predefined := desugerPreDefinedType(name); predefined != "" {
		return
	}

	if g.addFunction(name) {
		return
	}

	vi := findType(name)
	if vi != nil {
		g.addType(vi)
		return
	}

	ii := findInterface(name)
	if ii != nil {
		g.addInterface(ii)
		return
	}
	panicIf(true, "Didn't find function or variable with name '%s'", name)
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

func (g *goGenerator) generateTypeNamed(tp string) {
	// for e.g. pointer-to-struct types, we need to undo pointer-ness
	for tp[0] == '*' {
		tp = tp[1:]
	}

	// we assume this is array type like '[32]WCHAR', and we don't want
	// types for those
	if tp[0] == '[' && strings.Contains(tp, "]") {
		return
	}

	if s := desugerPreDefinedType(tp); s != "" {
		// skip pre-defined types
		return
	}

	vi := findType(tp)
	panicIf(vi == nil, "didn't find info about type '%s'\n", tp)

	if vi.WasGenerated {
		return
	}
	vi.WasGenerated = true

	v := vi.Variable
	if v.Type == typeAlias {
		v := vi.Variable
		g.ws("type %s %s\n", v.Name, v.Base)
		return
	}

	if v.Type == typeStruct {
		v := vi.Variable
		// first a pass to generate type names
		for _, f := range v.Field {
			tp := g.desugarTypeNamed(f.Type)
			g.generateTypeNamed(tp)
		}

		g.ws("type %s struct {\n", v.Name)
		for _, f := range v.Field {
			name := makeNameGoPublic(f.Name)
			tp := g.desugarTypeNamed(f.Type)
			g.ws("%s %s\n", name, tp)
		}
		g.ws("}\n\n")
		return
	}

	panicIf(true, "Unsupported type: '%s'", v.Type)
}

func (g *goGenerator) generateConsts(set []*Set) bool {
	// TODO: optimize if only one value
	if len(set) == 0 {
		return false
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
	return true
}

func (g *goGenerator) generateSet(vi *TypeInfo) {
	// TODO: can have Flag and Enum and Set in same Variable?
	if vi.WasGenerated {
		return
	}

	v := vi.Variable
	vi.WasGenerated = g.generateConsts(v.Set)
}

func makeNameGoPublic(s string) string {
	return strings.ToUpper(s[:1]) + s[1:]
}

func (g *goGenerator) generateEnum(vi *TypeInfo) {
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

func (g *goGenerator) generateFlag(vi *TypeInfo) {
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
		g.ws("\t%s *windows.LazyProc\n", fi.GoVarName())
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
		g.ws("\t%s = lib%s.NewProc(\"%s\")\n", fi.GoVarName(), dllNameNoExt, fi.Name)
	}

	g.ws("}\n")

	for _, fi := range mi.functions {
		g.generateFunction(fi)
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

func (g *goGenerator) desugarTypeNamed(tp string) string {
	if predefined := desugerPreDefinedType(tp); predefined != "" {
		return predefined
	}

	vi := findType(tp)
	panicIf(vi == nil, "didn't find info about type '%s'", tp)
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
	case "WCHAR":
		return tp
	case "UINT_PTR":
		return "uintptr"
	case "int":
		// yes, it's strange but that's what it seems to be
		return "int32"
	case "ModuleHandle":
		return "HANDLE"
	}
	if strings.ToLower(tp) == "void" {
		return "void"
	}
	return ""
}

// this converts type to a real Go type we want to use
func (g *goGenerator) desugarType(vi *TypeInfo) string {
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

	if tp == typeStruct {
		return v.Name
	}

	if tp == typeArray {
		base := g.desugarTypeNamed(v.Base)
		return fmt.Sprintf("[%s]%s", v.Count, base)
	}

	if tp == typeVoid {
		return "void"
	}

	if tp == typeUnion {
		panic("union NYI")
	}

	panicIf(true, "Unknown type '%s'", tp)

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

func (g *goGenerator) desugarReturnType(typeName string) string {
	// TODO: to handle BOOL => bool need a version of desugarTypeNamed()
	// specialized for return types
	if typeName == "BOOL" {
		return "bool"
	}
	returnType := g.desugarTypeNamed(typeName)
	g.generateTypeNamed(returnType)
	if strings.ToLower(returnType) == "void" {
		return ""
	}
	return returnType
}

func (g *goGenerator) generateFunction(fi *FunctionInfo) {
	fn := fi.Function

	// first a pass to generate types this function depends on
	for _, arg := range fn.Params {
		g.desugarTypeNamed(arg.Type)
		g.generateTypeNamed(arg.Type)
	}

	returnType := g.desugarReturnType(fn.Return.Type)

	g.ws("\nfunc %s(", fi.Name)
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
	g.ws(" " + returnType)
	g.ws(" {\n")

	// 	ret, _, _ := syscall.Syscall12(createWindowEx.Addr(), 12,
	nArgs := len(fn.Params)
	sysName, sysArgsCount := syscallFuncForArgCount(nArgs)
	if returnType != "" {
		g.ws("ret, _, _ := ")
	} else {
		g.ws("_, _, _ = ")
	}
	g.ws("%s(%s.Addr(), %d,\n", sysName, fi.GoVarName(), nArgs)
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
		nLeftOver--
	}
	g.ws(")\n")
	if returnType != "" {
		if returnType == "bool" {
			g.ws("return ret != 0\n")
		} else {
			g.ws("return %s(ret)\n", returnType)
		}
	}

	g.ws("\n}\n")
}

func (g *goGenerator) generateInterface(name string) {
	ii := findInterface(name)
	if ii.WasGenerated {
		return
	}
	i := ii.Interface
	baseName := i.BaseInterface
	if baseName != "" {
		g.generateInterface(baseName)
	}

	// TODO: generate vtable
	// TODO: generate functions wrapping vtable
	// TODO: generate function to create the type
	panic("NYI")

	ii.WasGenerated = true
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
	for _, name := range g.interfaces {
		g.generateInterface(name)
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
	fmt.Printf("Built index in %s. %d functions, %d variables, %d interfaces\n", time.Since(timeStart), len(allFunctions), len(allTypes), len(allInterfaces))

	g := newGoGenerator()
	g.addFunctionMust("CreateWindowExW")
	//g.addSymbol("FileTimeToSystemTime")
	//g.addSymbol("TzSpecificLocalTimeToSystemTime")
	//g.addSymbol("GetSystemTimeAsFileTime")
	//g.addSymbol("IBindHost")
	g.generate()
}
