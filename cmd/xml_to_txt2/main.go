package main

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/kjk/winapigen/pkg/xmldef"
)

type ParsedXmlFile struct {
	Name string
	Data *xmldef.ApiMonitor
}

var (
	parsedFiles []*ParsedXmlFile
	out         io.Writer
)

func must(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func panicIf(cond bool, format string, args ...interface{}) {
	if cond {
		err := fmt.Errorf(format, args...)
		must(err)
	}
}

func outf(format string, args ...interface{}) {
	fmt.Fprintf(out, format, args...)
}

func readZipFile(zf *zip.File) []byte {
	f, err := zf.Open()
	must(err)
	var buf bytes.Buffer
	_, err = io.Copy(&buf, f)
	must(err)
	return buf.Bytes()
}

func parseXmlDef(zf *zip.File) *xmldef.ApiMonitor {
	d := readZipFile(zf)
	var res xmldef.ApiMonitor
	err := xml.Unmarshal(d, &res)
	must(err)
	return &res
}

func parseApiMonitorData() {
	fileName := "api.zip"
	s, err := os.Stat(fileName)
	must(err)
	fileSize := s.Size()

	f, err := os.Open(fileName)
	must(err)
	defer f.Close()
	zr, err := zip.NewReader(f, fileSize)
	must(err)

	for _, fi := range zr.File {
		if fi.FileInfo().IsDir() {
			continue
		}
		data := parseXmlDef(fi)
		parsedFile := &ParsedXmlFile{
			Name: fi.Name,
			Data: data,
		}
		parsedFiles = append(parsedFiles, parsedFile)
	}
}

func serInclude(i *xmldef.Include) {
	outf("include %s\n", i.Filename)
}

type Args struct {
	a []string
}

func NewArgs(s ...string) *Args {
	return &Args{
		a: s,
	}
}

func escIfNeeded(s string) string {
	if strings.Contains(s, " ") {
		panicIf(strings.Contains(s, `"`), "can't escape '%s'", s)
		return `"` + s + `"`
	}
	return s
}

func (a *Args) AddNotEmpty(name, value string) {
	if value == "" {
		return
	}
	s := fmt.Sprintf("%s=%s", name, escIfNeeded(value))
	a.a = append(a.a, s)
}

func (a *Args) AddBool(name, value string) {
	if value == "" {
		return
	}
	panicIf(!strings.EqualFold(value, "true"), "value is '%s' and should be 'true'", value)
	a.a = append(a.a, name)
}

func (a *Args) String() string {
	return strings.Join(a.a, " ")
}

func addDisplay(args *Args, d *xmldef.Display) {
	if d == nil {
		return
	}
	args.AddNotEmpty("display", d.Name)
}

func serField(f *xmldef.Field) {
	args := NewArgs(f.Name, escIfNeeded(f.Type))
	args.AddNotEmpty("display", f.Display)
	args.AddNotEmpty("count", f.Count)
	args.AddNotEmpty("length", f.Length)
	args.AddNotEmpty("postCount", f.PostCount)
	args.AddNotEmpty("postLength", f.PostLength)
	args.AddNotEmpty("derefCount", f.DerefCount)
	args.AddBool("outputOnly", f.OutputOnly)
	outf("  %s\n", args.String())
}

func serVariable(v *xmldef.Variable) {
	panicIf(v.Success != nil, "Unexpected Sucess for %#v", v)

	tp := strings.ToLower(v.Type)
	if tp == "alias" {
		if v.Enum != nil {
			tp = "enum"
		}
		if v.Flag != nil {
			tp = "flag"
		}
	}
	args := NewArgs(tp, escIfNeeded(v.Name))

	switch tp {
	case "pointer", "alias", "enum", "flag":
		panicIf(v.Base == "", "missing Variable.Base for %#v", v)
		args.a = append(args.a, escIfNeeded(v.Base))
	case "array":
		args.AddNotEmpty("base", v.Base)
	default:
		panicIf(v.Base != "", "unexpected Variable.Base for %#v", v)
	}

	args.AddNotEmpty("count", v.Count)
	args.AddNotEmpty("pack", v.Pack)
	args.AddNotEmpty("pack32", v.Pack32)
	args.AddNotEmpty("pack64", v.Pack64)
	args.AddNotEmpty("size", v.Size)
	args.AddBool("unsigned", v.Unsigned)
	args.AddBool("displayhex", v.DisplayHex)
	addDisplay(args, v.Display)
	addFlag(args, v.Flag)
	addEnum(args, v.Enum)

	outf("%s\n", args.String())

	if v.Enum != nil {
		panicIf(v.Flag != nil, "both v.Enum and v.Flag set in %#v", v)
		serSet(v.Enum.Set)
	}

	if v.Flag != nil {
		panicIf(v.Enum != nil, "both v.Enum and v.Flag set %#v", v)
		serSet(v.Flag.Set)
	}

	switch tp {
	case "struct", "union":
		panicIf(len(v.Field) == 0, "no fields in %#v", v)
		serFields(v.Field)
	default:
		panicIf(len(v.Field) > 0, "unexpected fields in %#v", v)
	}
	outf("\n")
}

func serFields(fields []*xmldef.Field) {
	for _, f := range fields {
		serField(f)
	}
}

func serSet(set []*xmldef.Set) {
	maxLen := 0
	for _, s := range set {
		if len(s.Name) > maxLen {
			maxLen = len(s.Name)
		}
	}
	for _, s := range set {
		nSpaces := maxLen - len(s.Name)
		spaces := strings.Repeat(" ", nSpaces)
		fmt.Printf("  %s %s%s\n", s.Name, spaces, s.Value)
	}
}

func addEnum(args *Args, e *xmldef.Enum) {
	if e == nil {
		return
	}
	args.AddNotEmpty("defaultname", e.DefaultName)
	args.AddBool("reset", e.Reset)
}

func addFlag(args *Args, f *xmldef.Flag) {
	if f == nil {
		return
	}
	args.AddBool("reset", f.Reset)
}

func serCondition(c *xmldef.Condition) {
	outf("arch %s\n", c.Architecture)
	for _, v := range c.Variable {
		serVariable(v)
	}
	outf("arch off\n")
}

func serConditions(a []*xmldef.Condition) {
	for _, c := range a {
		serCondition(c)
	}
}

func serHeaders(h *xmldef.Headers) {
	if h == nil {
		return
	}
	outf("header\n")
	serConditions(h.Condition)
	serVariables(h.Variable)
}

func serVariables(vars []*xmldef.Variable) {
	for _, v := range vars {
		serVariable(v)
	}
}

func serIncludes(includes []*xmldef.Include) {
	if len(includes) == 0 {
		return
	}
	for _, i := range includes {
		serInclude(i)
	}
	outf("\n")
}

/* TODO:
Category           string          `xml:"Category,attr"  json:",omitempty"`
ErrorFunc          string          `xml:"ErrorFunc,attr"  json:",omitempty"`
OnlineHelp         string          `xml:"OnlineHelp,attr"  json:",omitempty"`

Categories         []*Category     `xml:"Category,omitempty" json:"Category,omitempty"`
ErrorDecode        []*ErrorDecode  `xml:"ErrorDecode,omitempty" json:"ErrorDecode,omitempty"`
ModuleAlias        []*ModuleAlias  `xml:"ModuleAlias,omitempty" json:"ModuleAlias,omitempty"`
SourceModule       []*SourceModule `xml:"SourceModule,omitempty" json:"SourceModule,omitempty"`
*/
func serModule(m *xmldef.Module) {
	args := NewArgs("dll", escIfNeeded(m.Name))
	args.AddNotEmpty("callingConvention", m.CallingConvention)
	args.AddBool("errorIsReturnValue", m.ErrorIsReturnValue)
	outf("%s\n", args.String())
	serConditions(m.Condition)
	serVariables(m.Variable)
	serApis(m.Api)
}

func serApis(apis []*xmldef.Api) {
	for _, api := range apis {
		serApi(api)
	}
}

/*
	ErrorFunc          string   `xml:"ErrorFunc,attr"  json:",omitempty"`
	ErrorIsReturnValue string   `xml:"ErrorIsReturnValue,attr"  json:",omitempty"`

	Param              []*Param `xml:"Param,omitempty" json:"Param,omitempty"`
	Return             *Return  `xml:"Return,omitempty" json:"Return,omitempty"`
	Success            *Success `xml:"Success,omitempty" json:"Success,omitempty"`
*/
func serApi(api *xmldef.Api) {
	args := NewArgs("func", escIfNeeded(api.Name))
	args.AddNotEmpty("display", api.Display)
	args.AddBool("bothCharset", api.BothCharset)

	args.AddNotEmpty("ordinal", api.Ordinal)
	args.AddNotEmpty("ordinalA", api.OrdinalA)
	args.AddNotEmpty("ordinalW", api.OrdinalW)

	// in vast majority, name of non-Unicode function is api.Name + "A".
	// Sometimes it's api.Name which is indicated by presenence of attribute
	// SuffixA with empty string as a value
	if api.SuffixA != nil {
		panicIf(*api.SuffixA != "", "unsupported SuffixA='%s' in %#v", *api.SuffixA, api)
		args.AddNotEmpty("nameA", api.Name)
	}

	args.AddBool("discard", api.Discard)
	args.AddBool("varArgs", api.VarArgs)
	args.AddNotEmpty("maxVarArgs", api.MaxVarArgs)
	outf("%s\n", args.String())
	// TODO: Return, Param, Success
}

func toTxt(f *ParsedXmlFile) {
	outf("File: %s\n", f.Name)

	d := f.Data
	// goes first as it's logically needed first
	serIncludes(d.Include)

	// ignore d.ApiSetSchema
	// ignore d.ErrorLookupModule
	serHeaders(d.Headers)
	// ignore d.HelpUrl
	// d.Include was serialized above
	//serInterface(d.Interface)
	for _, m := range d.Module {
		serModule(m)
	}
	// TODO: ensure exclusivity i.e. if Module then no Interface or Headers
}

func allToTxt() {
	nShown := 0
	for _, f := range parsedFiles {
		if nShown >= 3 {
			continue
		}
		if strings.Contains(f.Name, "Headers") {
			continue
		}
		if strings.Contains(f.Name, "Interfaces") {
			continue
		}
		if strings.Contains(f.Name, "Internal") {
			continue
		}
		if strings.Contains(f.Name, "MAPI") {
			continue
		}
		if strings.Contains(f.Name, "Microsoft.NET") {
			continue
		}
		if strings.Contains(f.Name, "MMF") {
			continue
		}
		if strings.Contains(f.Name, "Mozilla") {
			continue
		}
		if strings.Contains(f.Name, "SMI") {
			continue
		}
		if strings.Contains(f.Name, "VSS") {
			continue
		}
		if strings.Contains(f.Name, "WindowsFirewall") {
			continue
		}
		if strings.Contains(f.Name, "WindowsStore") {
			continue
		}
		toTxt(f)
		nShown++
	}
}

func main() {
	out = os.Stdout
	timeStart := time.Now()
	parseApiMonitorData()
	fmt.Printf("Parsed %d files in %s\n", len(parsedFiles), time.Since(timeStart))

	allToTxt()
}
