package main

import "strings"

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

func indexVariable(f *APIMonitorXMLFile, hdrs *Headers, mod *Module, cond *Condition, v *Variable) {
	if v.Type == typeInterface {
		// this only records to which file to attribute this interface
		name := v.Name
		ii := allInterfaces[name]
		if ii != nil {
			// must have been created when indexing <Interface> node
			// record FileName
			// Note: there are duplicates for the same. Too many to make
			// whitelist, so we let it slide for all of them and use
			// whatever the last name is
			ii.FileName = f.FileName
			return
		}

		ii = &InterfaceInfo{
			FileName: f.FileName,
		}
		allInterfaces[name] = ii
		return
	}

	{
		name := v.Name
		vi := &TypeInfo{
			SourceFile: f,
			Headers:    hdrs,
			Module:     mod,
			Condition:  cond,
			Variable:   v,
		}
		a := allTypes[name]
		allTypes[name] = append(a, vi)
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

func indexInterface(f *APIMonitorXMLFile) {
	if f.Interface == nil {
		return
	}
	name := f.Interface.Name
	ii := allInterfaces[name]
	if ii != nil {
		// must have been created when indexing a variable with Type Interfaces
		panicIf(ii.Interface != nil, "ii.Interface is not nil")
		ii.Interface = f.Interface
		return
	}

	ii = &InterfaceInfo{
		Interface: f.Interface,
	}
	allInterfaces[name] = ii
}

func buildIndex(files []*APIMonitorXMLFile) {
	for _, f := range files {
		if shouldSkipFile(f.FileName) {
			continue
		}
		indexModules(f)
		indexHeaders(f)
		indexInterface(f)
	}
}
