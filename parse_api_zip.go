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
	Name        string
	Type        string
	OutputOnly  bool
	PostCount   string
	Count       string
	InterfaceId string
}

// Return represents info parsed from:
// <Return Type="HRESULT" />
type Return struct {
	Type string
}

// API describes info parsed from <Api>
type API struct {
	Name    string
	Params  []Param
	Return  Return
	Success *Success
}

// Variable represents info parsed from:
//  <Variable Name=AsyncIBackgroundCopyCallback Type=Interface/>
type Variable struct {
	Name string
	Type string
	Base string
}

// Interface describes info parsed from:
// <Interface Name="AsyncIBackgroundCopyCallback" Id="{ca29d251-b4bb-4679-a3d9-ae8006119d54}" BaseInterface="IUnknown" OnlineHelp="MSDN" ErrorFunc="HRESULT" Category="Data Access and Storage/Background Intelligent Transfer Service (BITS)">
type Interface struct {
	Name          string
	BaseInterface string
	Id            string
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
	Variables         Variables
	Apis              []*API
}

// ParsedXML represents information extracted from a single .xml file
type ParsedXML struct {
	// name of the .xml file from which we extracted the data
	Path string
	// info from <Include> tags
	Includes      []string
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
	attr, rest := extractAttrByName(n.Attrs, "Filename")
	p.Includes = append(p.Includes, attr.Value)
	panicIf(len(rest) > 0, "<Include> has unsupported attributes. Node: '%s'", n)
}

func (p *parser) parseSuccess(n *XMLNode) *Success {
	mustTag(n.XMLName.Local, "Success")
	mustNoChildren(n)

	var res Success
	attrs := n.Attrs
	res.Return, attrs = mustExtractStringAttr(attrs, "Return", n)
	res.Value, attrs = mustExtractStringAttr(attrs, "Value", n)
	res.ErrorFunc, attrs = extractStringAttr(attrs, "ErrorFunc", "")
	mustNoAttrs(attrs, n)
	return &res
}

func (p *parser) parseSetValue(n *XMLNode) SetValue {
	attrs := n.Attrs
	name, attrs := mustExtractStringAttr(attrs, "Name", n)
	value, attrs := mustExtractStringAttr(attrs, "Value", n)
	mustNoAttrs(attrs, n)
	return SetValue{
		Name:  name,
		Value: value,
	}
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

	attrs := n.Attrs
	reset, attrs := extractBoolAttr(attrs, "Reset", false)
	defaultName, attrs := extractStringAttr(attrs, "DefaultName", "")
	mustNoAttrs(attrs, n)

	set := p.parseSet(n)
	e := Enum{
		Values:      set,
		Reset:       reset,
		DefaultName: defaultName,
	}
	return &e
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

func (p *parser) parseAlias(n *XMLNode, attrs []xml.Attr) Alias {
	mustTag(n.XMLName.Local, "Variable")
	var res Alias
	res.Name, attrs = mustExtractStringAttr(attrs, "Name", n)
	res.Base, attrs = mustExtractStringAttr(attrs, "Base", n)

	// must special-case this, should probably be wrapped inside Enum
	if res.Name == "FOURCC" {
		mustNoAttrs(attrs, n)
		set := p.parseSet(n)
		res.Enum = &Enum{
			Values: set,
		}
		return res
	}

	res.DisplayHex, attrs = extractBoolAttr(attrs, "DisplayHex", false)
	mustNoAttrs(attrs, n)

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

func (p *parser) parsePointer(n *XMLNode, attrs []xml.Attr) Pointer {
	mustTag(n.XMLName.Local, "Variable")
	var res Pointer
	res.Name, attrs = mustExtractStringAttr(attrs, "Name", n)
	res.Base, attrs = mustExtractStringAttr(attrs, "Base", n)
	mustNoAttrs(attrs, n)
	return res
}

func (p *parser) parseInteger(n *XMLNode, attrs []xml.Attr) Integer {
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
	res.Name, attrs = mustExtractStringAttr(attrs, "Name", n)
	res.Size, attrs = mustExtractIntAttr(attrs, "Size", n)
	res.Unsigned, attrs = extractBoolAttr(attrs, "Unsigned", false)
	res.DisplayHex, attrs = extractBoolAttr(attrs, "DisplayHex", false)
	mustNoAttrs(attrs, n)
	return res
}

func (p *parser) parseArray(n *XMLNode, attrs []xml.Attr) Array {
	mustTag(n.XMLName.Local, "Variable")
	var res Array
	res.Name, attrs = mustExtractStringAttr(attrs, "Name", n)
	res.Base, attrs = mustExtractStringAttr(attrs, "Base", n)
	res.Count, attrs = mustExtractIntAttr(attrs, "Count", n)
	mustNoAttrs(attrs, n)
	return res
}

func (p *parser) parseField(n *XMLNode) Field {
	mustNoChildren(n)
	attrs := n.Attrs
	typ, attrs := mustExtractStringAttr(attrs, "Type", n)
	name, attrs := mustExtractStringAttr(attrs, "Name", n)
	display, attrs := extractStringAttr(attrs, "Display", "")
	length, attrs := extractStringAttr(attrs, "Length", "")
	postLength, attrs := extractStringAttr(attrs, "PostLength", "")
	count, attrs := extractStringAttr(attrs, "Count", "")
	outputOnly, attrs := extractBoolAttr(attrs, "OutputOnly", false)
	derefCount, attrs := extractStringAttr(attrs, "DerefCount", "")
	mustNoAttrs(attrs, n)

	return Field{
		Type:       typ,
		Name:       name,
		Display:    display,
		Length:     length,
		PostLength: postLength,
		Count:      count,
		OutputOnly: outputOnly,
		DerefCount: derefCount,
	}
}

func (p *parser) parseDisplay(n *XMLNode) string {
	mustTag(n.XMLName.Local, "Display")
	mustNoChildren(n)
	display, attrs := mustExtractStringAttr(n.Attrs, "Name", n)
	mustNoAttrs(attrs, n)
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

func (p *parser) parseStruct(n *XMLNode, attrs []xml.Attr) Struct {
	mustTag(n.XMLName.Local, "Variable")
	var res Struct
	res.Name, attrs = mustExtractStringAttr(attrs, "Name", n)
	res.Pack, attrs = extractIntAttr(attrs, "Pack", -1)
	res.Pack32, attrs = extractIntAttr(attrs, "Pack32", -1)
	res.Pack64, attrs = extractIntAttr(attrs, "Pack64", -1)

	mustNoAttrs(attrs, n) // TODO: handle more attributes
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

func (p *parser) parseUnion(n *XMLNode, attrs []xml.Attr) Union {
	mustTag(n.XMLName.Local, "Variable")
	var res Union
	res.Name, attrs = mustExtractStringAttr(attrs, "Name", n)
	res.Pack, attrs = extractIntAttr(attrs, "Pack", -1)

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
	attrs := n.Attrs
	var res Param
	res.Name, attrs = mustExtractStringAttr(attrs, "Name", n)
	res.Type, attrs = mustExtractStringAttr(attrs, "Type", n)
	res.OutputOnly, attrs = extractBoolAttr(attrs, "OutputOnly", false)
	res.PostCount, attrs = extractStringAttr(attrs, "PostCount", "")
	res.Count, attrs = extractStringAttr(attrs, "Count", "")
	res.InterfaceId, attrs = extractStringAttr(attrs, "InterfaceId", "")
	mustNoAttrs(attrs, n)
	return res
}

func (p *parser) parseReturn(n *XMLNode) Return {
	mustTag(n.XMLName.Local, "Return")
	mustNoChildren(n)
	attrs := n.Attrs
	var res Return
	res.Type, attrs = mustExtractStringAttr(attrs, "Type", n)
	mustNoAttrs(attrs, n)
	return res
}

func (p *parser) parseAPIVariable(n *XMLNode) Variable {
	var res Variable
	attrs := n.Attrs
	res.Name, attrs = mustExtractStringAttr(attrs, "Name", n)
	res.Type, attrs = mustExtractStringAttr(attrs, "Type", n)
	res.Base, attrs = extractStringAttr(attrs, "Base", "")
	mustNoAttrs(attrs, n)
	return res
}

func (p *parser) parseAPI(n *XMLNode) API {
	mustTag(n.XMLName.Local, "Api")
	attrs := n.Attrs
	var res API
	res.Name, attrs = mustExtractStringAttr(attrs, "Name", n)
	mustNoAttrs(attrs, n)

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
	var res Interface
	attrs := n.Attrs

	res.Name, attrs = mustExtractStringAttr(attrs, "Name", n)
	res.BaseInterface, attrs = mustExtractStringAttr(attrs, "BaseInterface", n)
	res.Id, attrs = mustExtractStringAttr(attrs, "Id", n)
	res.OnlineHelp, attrs = extractStringAttr(attrs, "OnlineHelp", "")
	res.ErrorFunc, attrs = extractStringAttr(attrs, "ErrorFunc", "")
	res.Category, attrs = extractStringAttr(attrs, "Category", "")

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
	attrs := n.Attrs

	typ, attrs := mustExtractStringAttr(attrs, "Type", n)

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
		//mustNoAttrs(attrs, n)
		// not doing anything with those

	// some base types that don't need to exist
	case "Guid":
		// Note: not sure what to do
		mustNoChildren(n)
		//mustNoAttrs(attrs, n)
		//panicIf(name != "GUID", "unexpected name: '%s'", name)
	case "Void":
		// Note: not sure what to do
		mustNoChildren(n)
		//mustNoAttrs(attrs, n)
		//panicIf(name != "void", "unexpected name: '%s'", name)
	case "ModuleHandle":
		// Note: not sure what to do
		mustNoChildren(n)
		//mustNoAttrs(attrs, n)
		//panicIf(name != "HMODULE", "unexpected name: '%s'", name)
	case "Character":
		// Note: not sure what to do
		mustNoChildren(n)
		//mustNoAttrs(attrs, n)
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

func (p *parser) parseCondition(n *XMLNode) {
	mustTag(n.XMLName.Local, "Condition")
	attrs := n.Attrs
	arch, attrs := mustExtractIntAttr(attrs, "Architecture", n)
	mustNoAttrs(attrs, n)

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

func (p *parser) parseAPIMonitor(n *XMLNode) {
	mustTag(n.XMLName.Local, "ApiMonitor")
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
		case "HelpUrl":
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
	maxToProcess := 199
	nProcessed := 0
	for _, fi := range zr.File {
		if fi.FileInfo().IsDir() {
			//fmt.Printf("%s skipping because is dir\n", fi.Name)
			continue
		}
		fmt.Printf("%s\n", fi.Name)
		_, err := parseXML(fi)
		must(err)
		nProcessed++
		if nProcessed >= maxToProcess {
			break
		}
	}
}
