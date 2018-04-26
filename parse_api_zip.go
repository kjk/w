package main

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
)

// extracts definifions from api.zip which comes from api monitor
// http://www.rohitab.com/apimonitor

// Alias represents info parsed from <Variable Type="Alias">
// <Variable Name="LPCONTEXT" Type="Alias" Base="LPVOID" />
type Alias struct {
	Name       string
	Base       string
	Display    string
	DisplayHex bool
	Success    *Success
	Flag       *Flag
	Enum       *Enum
}

// Success represents info parsed from:
// <Success Return=NotEqual Value=-1/>
type Success struct {
	Return string // TODO: maybe use an enum? I assume there are only few valid values.
	//Value     int    // Note: not sure if those are always integers
	Value     string // TODO: sometimes the value is hex like 0xff
	ErrorFunc string
}

// Integer represents info parsed from:
// <Variable Name=PVOID64 Type=Integer Size=8 Unsigned=True DisplayHex=True/>
type Integer struct {
	Name       string
	Size       int
	Unsigned   bool
	DisplayHex bool
}

// SetValue represents a single flag in a set of values
// <Set Name="D3DDECLUSAGE_POSITION" Value="0" />
type SetValue struct {
	Name  string
	Value string
}

// Flag represents infor parsed from
// <Variable Name="D3DDECLUSAGE" Type="Alias" Base="BYTE">
// 	<Display Name="BYTE" />
// 	<Flag>
type Flag struct {
	Values []SetValue
}

// Enum represents infor parsed from
// <Variable Name="BTH_ADDR" Type="Alias" Base="ULONGLONG">
// <Enum>
//	<Set Name="BLUETOOTH_NULL_ADDRESS" Value="0" />
type Enum struct {
	Reset       bool // TODO: not sure what this means
	DefaultName string
	Values      []SetValue
}

// Pointer represents info parsed from <Variable Type="Pointer">
// <Variable Name="PCONTEXT_EX*" Type="Pointer"  Base="PCONTEXT_EX" />
type Pointer struct {
	Name string
	Base string
}

// Array represents info parsed from:
// <Variable Name="WCHAR [BLUETOOTH_MAX_NAME_SIZE]" Type="Array" Base="WCHAR" Count="248" />
type Array struct {
	Name  string
	Base  string
	Count int
}

// Field represents info parsed from:
// <Field Type=LPVOID Name=rgpLowerQualityChainContext Display=PCCERT_CHAIN_CONTEXT*/>
type Field struct {
	Name       string
	Type       string
	Display    string
	PostLength string
	Length     string // this might refer to a field name
	Count      string // this might refer to a field name
	OutputOnly bool
	DerefCount string // this might refer to a field name
}

// Struct represents info parsed from:
// <Variable Name="BLUETOOTH_ADDRESS" Type="Struct">
type Struct struct {
	Name    string
	Display string
	Fields  []Field
	Pack    int
	Pack32  int
	Pack64  int
}

// Union represents info parsed from:
// <Variable Name="[SDP_ELEMENT_DATA_u]" Type="Union">
type Union struct {
	Name    string
	Display string
	Fields  []Field
	Pack    int
}

// Param represents info parsed from:
// <Param Type="IBackgroundCopyJob*" Name="pJob" />
type Param struct {
	Name            string
	Type            string
	Display         string
	OutputOnly      bool
	PostCount       string // refers to Name of other param
	PostLength      string // refers to Name of other param
	Count           string
	Length          string // might refer to Name of other param
	InterfaceID     string
	DerefPostCount  string
	DerefCount      string
	DerefPostLength string
}

// Return represents info parsed from:
// <Return Type="HRESULT" />
type Return struct {
	Type string
}

// API describes info parsed from <Api>
type API struct {
	Name             string
	Params           []Param
	VarArgs          bool
	Return           Return
	Success          *Success
	Ordinal          int
	OrdinalA         int
	OrdinalW         int
	BothCharsets     bool
	Discard          bool
	Disabled_Discard bool
}

// Variable represents info parsed from:
//  <Variable Name=AsyncIBackgroundCopyCallback Type=Interface/>
// <Variable Name="wchar_t [1]" Type=Array Base=wchar_t Count=1/>
type Variable struct {
	Name  string
	Type  string
	Base  string
	Count int // if Type is Array
	Pack  int // if Type is Struct
}

// Interface describes info parsed from:
// <Interface Name="AsyncIBackgroundCopyCallback" Id="{ca29d251-b4bb-4679-a3d9-ae8006119d54}" BaseInterface="IUnknown" OnlineHelp="MSDN" ErrorFunc="HRESULT" Category="Data Access and Storage/Background Intelligent Transfer Service (BITS)">
type Interface struct {
	Name          string
	BaseInterface string
	ID            string
	OnlineHelp    string
	ErrorFunc     string
	Category      string
	Methods       []API
	Variables     []Variable
}

// Variables describes info parsed from <Variable>
type Variables struct {
	Pointers   []Pointer
	Arrays     []Array
	Structs    []Struct
	Unions     []Union
	Aliases    []Alias
	Integers   []Integer
	Interfaces []Interface
}

// Module is a result of parsing <Module>
type Module struct {
	Name              string // e.g. "kernel32.dll"
	CallingConvention string // e.g. STDCALL, CDECL
	ErrorFunc         string // e.g. GetLastError
	OnlineHelp        string // e.g. MSDN
	Variables         *Variables
	Apis              []API
	Aliases           []string
}

// ParsedXML represents information extracted from a single .xml file
type ParsedXML struct {
	// name of the .xml file from which we extracted the data
	Path string
	// info from <Include> tags
	Includes      []string
	Modules       []*Module
	Variables     *Variables
	Arch32        *Variables
	Arch64        *Variables
	currVariables *Variables
}

// parser holds state for parsing a current file
type parser struct {
	ParsedXML
	root   *XMLNode
	indent int
	dec    *xml.Decoder
}

func newParser(fi *zip.File) (*parser, error) {
	res := &parser{}
	res.Variables = &Variables{}
	res.currVariables = res.Variables
	res.Path = fi.Name
	d, err := readZipFile(fi)
	if err != nil {
		return nil, err
	}
	r := bytes.NewBuffer(d)
	res.dec = xml.NewDecoder(r)
	return res, nil
}

func (p *parser) parseInclude(n *XMLNode) {
	mustTag(n.XMLName.Local, "Include")
	attrs := NewAttrs(n)
	fileName := attrs.mustExtractString("Filename")
	p.Includes = append(p.Includes, fileName)
	attrs.mustEmpty()
}

// <Success />
// <Success Return="NotEqual" Value="-1" />
func (p *parser) parseSuccess(n *XMLNode) *Success {
	mustTag(n.XMLName.Local, "Success")
	mustNoChildren(n)

	var res Success
	// TODO: not sure what <Success /> means, but they do happen
	if len(n.Attrs) == 0 {
		return &res
	}

	attrs := NewAttrs(n)
	res.Return = attrs.mustExtractString("Return")
	res.Value = attrs.mustExtractString("Value")
	res.ErrorFunc = attrs.extractString("ErrorFunc")
	attrs.mustEmpty()
	return &res
}

func (p *parser) parseSetValue(n *XMLNode) SetValue {
	var res SetValue
	attrs := NewAttrs(n)
	res.Name = attrs.mustExtractString("Name")
	res.Value = attrs.mustExtractString("Value")
	attrs.mustEmpty()
	return res
}

func (p *parser) parseSet(n *XMLNode) []SetValue {
	tag := n.XMLName.Local
	panicIf(!(tag == "Enum" || tag == "Flag" || tag == "Variable"), "tag is '%s'", tag)

	var set []SetValue
	for _, n := range n.Nodes {
		tag := n.XMLName.Local
		switch tag {
		case "Set":
			setValue := p.parseSetValue(n)
			set = append(set, setValue)
		default:
			panicIf(true, "unsupported child node: '%s'", n)
		}
	}
	return set
}

func (p *parser) parseEnum(n *XMLNode) *Enum {
	mustTag(n.XMLName.Local, "Enum")
	var res Enum

	attrs := NewAttrs(n)
	res.Reset = attrs.extractBool("Reset")
	res.DefaultName = attrs.extractString("DefaultName")
	attrs.mustEmpty()
	res.Values = p.parseSet(n)
	return &res
}

func (p *parser) parseFlag(n *XMLNode) *Flag {
	mustTag(n.XMLName.Local, "Flag")
	mustNoAttrs(n.Attrs, n)

	set := p.parseSet(n)
	f := Flag{
		Values: set,
	}
	return &f
}

func (p *parser) parseAlias(n *XMLNode, attrs *Attrs) Alias {
	mustTag(n.XMLName.Local, "Variable")
	var res Alias
	res.Name = attrs.mustExtractString("Name")
	res.Base = attrs.mustExtractString("Base")

	// must special-case this, should probably be wrapped inside Enum
	if res.Name == "FOURCC" {
		attrs.mustEmpty()
		set := p.parseSet(n)
		res.Enum = &Enum{
			Values: set,
		}
		return res
	}

	res.DisplayHex = attrs.extractBool("DisplayHex")
	attrs.mustEmpty()

	for _, n := range n.Nodes {
		tag := n.XMLName.Local
		switch tag {
		case "Enum":
			res.Enum = p.parseEnum(n)
		case "Flag":
			res.Flag = p.parseFlag(n)
		case "Display":
			res.Display = p.parseDisplay(n)
		case "Success":
			res.Success = p.parseSuccess(n)
		default:
			panicIf(true, "unsupported child node: '%s'", n)
		}
	}
	return res
}

func (p *parser) parsePointer(n *XMLNode, attrs *Attrs) Pointer {
	mustTag(n.XMLName.Local, "Variable")
	var res Pointer
	res.Name = attrs.mustExtractString("Name")
	res.Base = attrs.mustExtractString("Base")
	attrs.mustEmpty()
	return res
}

func (p *parser) parseInteger(n *XMLNode, attrs *Attrs) Integer {
	mustTag(n.XMLName.Local, "Variable")

	/* Note: this doesn't fully parse
	   <Variable Name=BOOL Type=Integer Size=4>
	     <Enum DefaultName=TRUE>
	     <Set Name=TRUE Value=1/>
	     <Set Name=FALSE Value=0/>
	   </Enum>
	   <Success Return=NotEqual Value=0/>
	   </Variable>
	*/
	//mustNoChildren(n)

	var res Integer
	res.Name = attrs.mustExtractString("Name")
	res.Size = attrs.mustExtractInt("Size")
	res.Unsigned = attrs.extractBool("Unsigned")
	res.DisplayHex = attrs.extractBool("DisplayHex")
	attrs.mustEmpty()
	return res
}

func (p *parser) parseArray(n *XMLNode, attrs *Attrs) Array {
	mustTag(n.XMLName.Local, "Variable")
	var res Array
	res.Name = attrs.mustExtractString("Name")
	res.Base = attrs.mustExtractString("Base")
	res.Count = attrs.mustExtractInt("Count")
	attrs.mustEmpty()
	return res
}

func (p *parser) parseField(n *XMLNode) Field {
	mustNoChildren(n)
	var res Field
	attrs := NewAttrs(n)
	res.Type = attrs.mustExtractString("Type")
	res.Name = attrs.mustExtractString("Name")
	res.Display = attrs.extractString("Display")
	res.Length = attrs.extractString("Length")
	res.PostLength = attrs.extractString("PostLength")
	res.Count = attrs.extractString("Count")
	res.OutputOnly = attrs.extractBool("OutputOnly")
	res.DerefCount = attrs.extractString("DerefCount")
	attrs.mustEmpty()
	return res
}

func (p *parser) parseDisplay(n *XMLNode) string {
	mustTag(n.XMLName.Local, "Display")
	mustNoChildren(n)
	attrs := NewAttrs(n)
	display := attrs.mustExtractString("Name")
	attrs.mustEmpty()
	return display
}

func validateFields(fields []Field) {
	names := map[string]bool{}
	for _, f := range fields {
		names[f.Name] = true
	}
	for _, f := range fields {
		if f.Length != "" {
			panicIf(!names[f.Length], "Length '%s' doesn't map to any field", f.Length)
		}
		if f.Count != "" {
			panicIf(!names[f.Count], "Count '%s' doesn't map to any field", f.Count)
		}
	}
}

func (p *parser) parseStruct(n *XMLNode, attrs *Attrs) Struct {
	mustTag(n.XMLName.Local, "Variable")
	var res Struct
	res.Name = attrs.mustExtractString("Name")
	res.Pack = attrs.extractInt("Pack", -1)
	res.Pack32 = attrs.extractInt("Pack32", -1)
	res.Pack64 = attrs.extractInt("Pack64", -1)

	attrs.mustEmpty()
	for _, n := range n.Nodes {
		tag := n.XMLName.Local
		switch tag {
		case "Field":
			field := p.parseField(n)
			res.Fields = append(res.Fields, field)
		case "Display":
			res.Display = p.parseDisplay(n)
		default:
			panicIf(true, "unsupported element %s", n)
		}
	}
	validateFields(res.Fields)
	return res
}

func (p *parser) parseUnion(n *XMLNode, attrs *Attrs) Union {
	mustTag(n.XMLName.Local, "Variable")
	var res Union
	res.Name = attrs.mustExtractString("Name")
	res.Pack = attrs.extractInt("Pack", -1)

	for _, n := range n.Nodes {
		tag := n.XMLName.Local
		switch tag {
		case "Field":
			field := p.parseField(n)
			res.Fields = append(res.Fields, field)
		case "Display":
			res.Display = p.parseDisplay(n)
		default:
			panicIf(true, "unsupported element %s", n)
		}
	}
	validateFields(res.Fields)
	return res
}

func (p *parser) parseParam(n *XMLNode) Param {
	mustTag(n.XMLName.Local, "Param")
	mustNoChildren(n)
	attrs := NewAttrs(n)
	var res Param
	res.Name = attrs.mustExtractString("Name")
	res.Type = attrs.mustExtractString("Type")
	res.OutputOnly = attrs.extractBool("OutputOnly")
	res.PostCount = attrs.extractString("PostCount")
	res.PostLength = attrs.extractString("PostLength")
	res.Count = attrs.extractString("Count")
	res.Length = attrs.extractString("Length")
	res.InterfaceID = attrs.extractString("InterfaceId")
	res.Display = attrs.extractString("Display")
	res.DerefPostCount = attrs.extractString("DerefPostCount")
	res.DerefPostLength = attrs.extractString("DerefPostLength")
	res.DerefCount = attrs.extractString("DerefCount")
	attrs.mustEmpty()
	return res
}

func (p *parser) parseReturn(n *XMLNode) Return {
	mustTag(n.XMLName.Local, "Return")
	mustNoChildren(n)
	attrs := NewAttrs(n)
	var res Return
	res.Type = attrs.mustExtractString("Type")
	attrs.mustEmpty()
	return res
}

func (p *parser) parseAPIVariable(n *XMLNode) Variable {
	var res Variable
	attrs := NewAttrs(n)
	res.Name = attrs.mustExtractString("Name")
	res.Type = attrs.mustExtractString("Type")
	res.Base = attrs.extractString("Base")
	res.Count = attrs.extractInt("Count", 0)
	res.Pack = attrs.extractInt("Pack", 0)
	attrs.mustEmpty()
	return res
}

// <Api Name=Output VarArgs=True>
func (p *parser) parseAPI(n *XMLNode) API {
	mustTag(n.XMLName.Local, "Api")
	attrs := NewAttrs(n)
	var res API
	res.Name = attrs.mustExtractString("Name")
	res.VarArgs = attrs.extractBool("VarArgs")
	res.Ordinal = attrs.extractInt("Ordinal", -1)
	res.OrdinalA = attrs.extractInt("OrdinalA", -1)
	res.OrdinalW = attrs.extractInt("OrdinalW", -1)
	res.BothCharsets = attrs.extractBool("BothCharset")
	res.Discard = attrs.extractBool("Discard")
	res.Disabled_Discard = attrs.extractBool("Disabled_Discard")
	attrs.mustEmpty()

	for _, n := range n.Nodes {
		tag := n.XMLName.Local
		switch tag {
		case "Param":
			param := p.parseParam(n)
			res.Params = append(res.Params, param)
		case "Return":
			res.Return = p.parseReturn(n)
		case "Success":
			res.Success = p.parseSuccess(n)
		default:
			must(fmt.Errorf("unuspported node:\n%s", n))
		}
	}

	return res
}

func (p *parser) parseInterface(n *XMLNode) Interface {
	mustTag(n.XMLName.Local, "Interface")
	attrs := NewAttrs(n)

	var res Interface
	res.Name = attrs.mustExtractString("Name")
	res.BaseInterface = attrs.extractString("BaseInterface")
	res.ID = attrs.extractString("Id")
	res.OnlineHelp = attrs.extractString("OnlineHelp")
	res.ErrorFunc = attrs.extractString("ErrorFunc")
	res.Category = attrs.extractString("Category")

	for _, n := range n.Nodes {
		tag := n.XMLName.Local
		switch tag {
		case "Api":
			api := p.parseAPI(n)
			res.Methods = append(res.Methods, api)
		case "Variable":
			v := p.parseAPIVariable(n)
			res.Variables = append(res.Variables, v)

		default:
			must(fmt.Errorf("unuspported node:\n%s", n))
		}
	}
	return res
}

func (p *parser) parseVariable(n *XMLNode) {
	mustTag(n.XMLName.Local, "Variable")
	attrs := NewAttrs(n)
	typ := attrs.mustExtractString("Type")

	switch typ {
	case "Alias":
		a := p.parseAlias(n, attrs)
		p.currVariables.Aliases = append(p.currVariables.Aliases, a)
	case "Pointer":
		ptr := p.parsePointer(n, attrs)
		p.currVariables.Pointers = append(p.currVariables.Pointers, ptr)
	case "Array":
		arr := p.parseArray(n, attrs)
		p.currVariables.Arrays = append(p.currVariables.Arrays, arr)
	case "Struct":
		s := p.parseStruct(n, attrs)
		p.currVariables.Structs = append(p.currVariables.Structs, s)
	case "Union":
		u := p.parseUnion(n, attrs)
		p.currVariables.Unions = append(p.currVariables.Unions, u)
	case "Integer":
		i := p.parseInteger(n, attrs)
		p.currVariables.Integers = append(p.currVariables.Integers, i)

	case "Interface":
		// Note: don't understand why I need those if we have <Interface> definitions
		mustNoChildren(n)
		//	attrs.mustEmpty()
		// not doing anything with those

	// some base types that don't need to exist
	case "Guid":
		// Note: not sure what to do
		mustNoChildren(n)
		//	attrs.mustEmpty()
		//panicIf(name != "GUID", "unexpected name: '%s'", name)
	case "Void":
		// Note: not sure what to do
		mustNoChildren(n)
		//attrs.mustEmpty()
		//panicIf(name != "void", "unexpected name: '%s'", name)
	case "ModuleHandle":
		// Note: not sure what to do
		mustNoChildren(n)
		//attrs.mustEmpty()
		//panicIf(name != "HMODULE", "unexpected name: '%s'", name)
	case "Character":
		// Note: not sure what to do
		mustNoChildren(n)
		//attrs.mustEmpty()
		//panicIf(name != "CHAR", "unexpected name: '%s'", name)
	case "UnicodeCharacter":
		// Note: not sure what to do
		mustNoChildren(n)
		//panicIf(name != "WCHAR", "unexpected name: '%s'", name)
	case "TCharacter":
		// Note: not sure what to do
		mustNoChildren(n)
		//panicIf(name != "TCHAR", "unexpected name: '%s'", name)
	case "Floating":
		// Note: not sure what to do
		mustNoChildren(n)
		//panicIf(!(name == "float" || name == "double"), "unexpected float name: '%s'", name)
	default:
		panicIf(true, "Unsupported Type attribute '%s'. Node:\n%s\n", typ, xmlSerializeNode(n, 0))
	}
}

func (p *parser) parseModuleAlias(n *XMLNode) string {
	mustTag(n.XMLName.Local, "ModuleAlias")
	mustNoChildren(n)
	attrs := NewAttrs(n)
	alias := attrs.mustExtractString("Name")
	attrs.mustEmpty()
	return alias
}

func (p *parser) parseCondition(n *XMLNode) {
	mustTag(n.XMLName.Local, "Condition")
	attrs := NewAttrs(n)
	arch := attrs.mustExtractInt("Architecture")
	attrs.mustEmpty()

	prevVariables := p.currVariables
	switch arch {
	case 32:
		if p.Arch32 == nil {
			p.Arch32 = &Variables{}
		}
		p.currVariables = p.Arch32
	case 64:
		if p.Arch64 == nil {
			p.Arch64 = &Variables{}
		}
		p.currVariables = p.Arch64
	default:
		must(fmt.Errorf("unknown Architecture '%d'", arch))
	}

	for _, n := range n.Nodes {
		tag := n.XMLName.Local
		switch tag {
		case "Variable":
			p.parseVariable(n)
		default:
			must(fmt.Errorf("unuspported node:\n%s", n))
		}
	}

	p.currVariables = prevVariables
}

func (p *parser) parseHeaders(n *XMLNode) {
	mustTag(n.XMLName.Local, "Headers")
	mustNoAttrs(n.Attrs, n)
	for _, n := range n.Nodes {
		tag := n.XMLName.Local
		switch tag {
		case "Variable":
			p.parseVariable(n)
		case "Condition":
			p.parseCondition(n)
		default:
			must(fmt.Errorf("unuspported node:\n%s", n))
		}
	}
}

// <Module Name=DbgEng.dll CallingConvention=STDCALL ErrorFunc=HRESULT OnlineHelp=MSDN>
func (p *parser) parseModule(n *XMLNode) *Module {
	mustTag(n.XMLName.Local, "Module")
	var res Module
	attrs := NewAttrs(n)
	res.Name = attrs.mustExtractString("Name")
	res.CallingConvention = attrs.mustExtractString("CallingConvention")
	res.ErrorFunc = attrs.mustExtractString("ErrorFunc")
	res.OnlineHelp = attrs.extractString("OnlineHelp")
	// discard Category
	_ = attrs.extractString("Category")
	res.Variables = &Variables{}

	prevVariables := p.currVariables
	p.currVariables = res.Variables
	attrs.mustEmpty()

	for _, n := range n.Nodes {
		tag := n.XMLName.Local
		switch tag {
		case "Variable":
			p.parseVariable(n)
		case "Category":
			mustNoChildren(n)
			// ignore Category for now
		case "Api":
			api := p.parseAPI(n)
			res.Apis = append(res.Apis, api)
		case "ModuleAlias":
			alias := p.parseModuleAlias(n)
			res.Aliases = append(res.Aliases, alias)
		default:
			must(fmt.Errorf("unuspported node:\n%s", n))
		}
	}

	p.currVariables = prevVariables
	return &res
}

func (p *parser) parseAPIMonitor(n *XMLNode) {
	mustTag(n.XMLName.Local, "ApiMonitor")
	mustNoAttrs(n.Attrs, n)

	for _, n := range n.Nodes {
		tag := n.XMLName.Local
		switch tag {
		case "Include":
			p.parseInclude(n)
		case "Headers":
			p.parseHeaders(n)
		case "Interface":
			i := p.parseInterface(n)
			p.currVariables.Interfaces = append(p.currVariables.Interfaces, i)
		case "Module":
			m := p.parseModule(n)
			p.Modules = append(p.Modules, m)
		case "HelpUrl":
			// ignore
		case "ErrorLookupModule":
			// TODO: ignore for now. Don't know what that means
			// used in IFilterGraph.xml
		case "ApiSetSchema":
			// ignore
		case "UnsupportedModules":
			// ignore
		default:
			must(fmt.Errorf("unuspported node:\n%s", n))
		}
	}
}

func (p *parser) parse() {
}

func parseXML(fi *zip.File) (*ParsedXML, error) {
	p, err := newParser(fi)
	if err != nil {
		return nil, err
	}
	var n XMLNode
	err = p.dec.Decode(&n)
	must(err)
	p.root = &n
	p.parseAPIMonitor(p.root)
	//xmlPrint(parser.root)
	return &p.ParsedXML, nil
}

func parseAPIZip() {
	fileName := "api.zip"
	st, err := os.Stat(fileName)
	must(err)
	f, err := os.Open(fileName)
	must(err)
	defer f.Close()
	zr, err := zip.NewReader(f, st.Size())
	must(err)
	for _, fi := range zr.File {
		if fi.FileInfo().IsDir() {
			//fmt.Printf("%s skipping because is dir\n", fi.Name)
			continue
		}
		fmt.Printf("%s\n", fi.Name)
		_, err := parseXML(fi)
		must(err)
	}
}
