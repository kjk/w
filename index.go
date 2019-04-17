package main

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
)

const (
	// names for Variable.Type field
	typeAlias        = "Alias"
	typePtr          = "Pointer"
	typePointer      = "Pointer"
	typeInterface    = "Interface"
	typeStruct       = "Struct"
	typeUnion        = "Union"
	typeArray        = "Array"
	typeVoid         = "Void"
	typeInteger      = "Integer"
	typeModuleHandle = "ModuleHandle"
)

// NameAndType describes an argument to a function
type NameAndType struct {
	Name     string
	TypeName string
}

// TypeInfo describes a type
type TypeInfo struct {
	SourceFile *APIMonitorXMLFile
	Headers    *Headers
	Module     *Module    // can be nil
	Condition  *Condition // can be nil
	Variable   *Variable

	// this is a name of the type for a given language e.g. Go
	// it might be different than the name of C type
	// after desugaring type (e.g. resolving LPWSTR => *uint16)
	Name string

	// for structs
	Fields []*NameAndType

	WasAdded     bool
	WasGenerated bool
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

	Args []*NameAndType
	// can be void if no return type
	ReturnType string

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

// InterfaceDeclarationInfo represents declaration of
// an interface in Headers section. There are multiple
// declaration per header file.
// E.g. IBindCtx in Headers\ole.h.xml
type InterfaceDeclarationInfo struct {
	SourceFile *APIMonitorXMLFile
	Headers    *Headers
	Module     *Module
	Name       string // e.g. IBindCtx
}

// InterfaceDefinitionInfo represents definition of an interface
// in an Interface section.
// E.g. IBindCtx in Interfaces\COM\IBindCtx.h.xml
type InterfaceDefinitionInfo struct {
	SourceFile *APIMonitorXMLFile
	Interface  *Interface
}

// InterfaceInfo describes an interface. It combines information about
// interfaces from 2 sources. E.g. IPropertySetStorage
// * declaration in a header file Headers\ole.h.xml file
//   in Headers section or in Interface.Variable section
// * definition in interfaces\COM\IBindCtx
type InterfaceInfo struct {
	Declaration *InterfaceDeclarationInfo
	Definition  *InterfaceDefinitionInfo

	// when interface is declared inside Interface
	// definition, we want to attribute it to the header
	// file of the parent interface
	Parent *InterfaceInfo

	// Interface.Api but with more info
	Methods []*FunctionInfo

	// Interface.Variable but with more info
	// see e.g. "IBitsPeer"
	Types []*TypeInfo

	// Interface.Variable when Type=Interface
	// see e.g. "IBitsPeer"
	Interfaces []*InterfaceInfo

	WasAdded     bool
	WasGenerated bool
}

// Name returns name of the interface
func (ii *InterfaceInfo) Name() string {
	return ii.Definition.Interface.Name
}

// e.g. gdi32.go
func (m *Module) goSourceFileName() string {
	return m.moduleName() + ".go"
}

// e.g. gdi32
func (m *Module) moduleName() string {
	s := strings.ToLower(m.Name)
	s = strings.TrimSuffix(s, ".dll")
	panicIf(filepath.Ext(s) != "", "unexpected m.Name '%s'", m.Name)
	return s
}

// e.g. gdi32.dll
func (m *Module) dllName() string {
	return m.moduleName() + ".dll"
}

func (ii *InterfaceInfo) goSourceFileName() string {
	// we want to attribute interfaces declared inside
	// interface definitions to the header file
	// that declared parent interface
	for ii.Parent != nil {
		ii = ii.Parent
	}
	d := ii.Declaration
	if d == nil {
		fmt.Printf("ii.Declaration for %s is nil\n", ii.Definition.Interface.Name)
	}
	panicIf(d == nil, "ii.Declaration is nil")
	// if have module, attribute interface to that module
	if d.Module != nil {
		return d.Module.goSourceFileName()
	}

	// Url.h.xml => url.h.go
	f := d.SourceFile
	name := filepath.ToSlash(f.FileName)
	name = path.Base(name)
	name = strings.ToLower(name)
	panicIf(!strings.HasSuffix(name, ".xml"), "'%s' doesn't end with .xml", f.FileName)
	name = strings.TrimSuffix(name, ".xml")
	return name + ".go"
}

var (
	// we might hvae functions with the same name in different libraries
	allFunctions  = map[string][]*FunctionInfo{}
	allTypes      = map[string][]*TypeInfo{}
	allInterfaces = map[string]*InterfaceInfo{}
)

func indexFunction(f *APIMonitorXMLFile, mod *Module, api *Api) {
	if api.BothCharset == "" {
		name := api.Name
		fi := &FunctionInfo{
			Function:   api,
			Name:       name,
			Module:     mod,
			SourceFile: f,
		}
		//panicIf(allFunctions[name] != nil, "allFunctions[%s] is already set", name)
		a := allFunctions[name]
		allFunctions[name] = append(a, fi)
		return
	}

	panicIf(api.BothCharset != "True", "api.BothCharsets is '%s'", api.BothCharset)

	{
		name := api.Name + "W"
		fi := &FunctionInfo{
			Function:   api,
			Name:       name,
			Module:     mod,
			SourceFile: f,
			IsUnicode:  true,
		}
		a := allFunctions[name]
		allFunctions[name] = append(a, fi)
	}

	{
		name := api.Name + "A"
		fi := &FunctionInfo{
			Function:   api,
			Name:       name,
			Module:     mod,
			SourceFile: f,
			IsUnicode:  false,
		}
		a := allFunctions[name]
		allFunctions[name] = append(a, fi)
	}
}

func indexVariableInterface(f *APIMonitorXMLFile, hdrs *Headers, mod *Module, cond *Condition, v *Variable) bool {
	if v.Type != typeInterface {
		return false
	}
	if mod != nil {
		//fmt.Printf("Interface '%s' declared in module '%s' in file '%s'\n", v.Name, mod.Name, f.FileName)
	}
	panicIf(cond != nil, "can cond be non-nil?")
	// this is an interface declaration that comes either
	// from Headers section or from Interface.Variables section
	d := &InterfaceDeclarationInfo{
		SourceFile: f,
		Headers:    hdrs,
		Name:       v.Name,
		Module:     mod,
	}
	ii := allInterfaces[v.Name]
	if ii == nil {
		ii = &InterfaceInfo{
			Declaration: d,
		}
		allInterfaces[v.Name] = ii
		return true
	}

	if ii.Declaration != nil {
		// TODO: maybe remember all files and prioritise for attribution
		// e.g. windows.h.xml over ole.h.xml over multimedia.h.xml
		// it happens, inform about it but let it slide
		//fmt.Printf("Duplicate declaration of interface '%s' in %s and %s\n", v.Name, ii.Declaration.SourceFile.FileName, f.FileName)
		return true
	}
	ii.Declaration = d
	return true
}

func indexVariable(f *APIMonitorXMLFile, hdrs *Headers, mod *Module, cond *Condition, v *Variable) {
	if indexVariableInterface(f, hdrs, mod, cond, v) {
		return
	}

	// panicIf(mod == nil, "module for %s in file %s is nil", v.Name, f.FileName)

	vi := &TypeInfo{
		SourceFile: f,
		Headers:    hdrs,
		Module:     mod,
		Condition:  cond,
		Variable:   v,
	}
	a := allTypes[v.Name]
	allTypes[v.Name] = append(a, vi)
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

func indexInterfaceDefinition(f *APIMonitorXMLFile) {
	if f.Interface == nil {
		return
	}
	i := f.Interface
	d := &InterfaceDefinitionInfo{
		SourceFile: f,
		Interface:  i,
	}
	name := i.Name
	ii := allInterfaces[name]
	if ii == nil {
		ii = &InterfaceInfo{
			Definition: d,
		}
		allInterfaces[name] = ii
		return
	}
	panicIf(ii.Definition != nil, "unexpected double definition")
	ii.Definition = d
	for _, v := range i.Variable {
		indexVariable(f, nil, nil, nil, v)
	}
}

func buildIndex(files []*APIMonitorXMLFile) {
	for _, f := range files {
		if shouldSkipFile(f.FileName) {
			continue
		}
		indexModules(f)
		indexHeaders(f)
		indexInterfaceDefinition(f)
	}
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
