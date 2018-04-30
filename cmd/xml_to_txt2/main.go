package main

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
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

func serParam(p *xmldef.Param, indent int) {
	args := NewArgs(p.Name, escIfNeeded(p.Type))
	args.AddNotEmpty("display", p.Display)
	args.AddNotEmpty("interfaceId", p.InterfaceId)
	args.AddNotEmpty("count", p.Count)
	args.AddNotEmpty("length", p.Length)
	args.AddNotEmpty("postCount", p.PostCount)
	args.AddNotEmpty("postLength", p.PostLength)
	args.AddNotEmpty("derefPostCount", p.DerefPostCount)
	args.AddNotEmpty("derefPostLength", p.DerefPostLength)
	args.AddNotEmpty("derefCount", p.DerefCount)
	args.AddBool("outputOnly", p.OutputOnly)
	indentStr := strings.Repeat("  ", indent)
	outf("%s%s\n", indentStr, args.String())
}

func serVariable(v *xmldef.Variable, indent int) {
	// fix what looks like a mistake in .xml definition
	if v.Set != nil {
		panicIf(v.Name != "FOURCC", "unsupported var with set for '%s'", v.Name)
		v.Enum = &xmldef.Enum{
			Set: v.Set,
		}
		v.Set = nil
	}

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
	args.AddBool("displayHex", v.DisplayHex)
	addDisplay(args, v.Display)
	addFlag(args, v.Flag)
	addEnum(args, v.Enum)

	indentStr := strings.Repeat("  ", indent)
	outf("%s%s\n", indentStr, args.String())

	serSuccess(v.Success, indent+1)

	if v.Enum != nil {
		panicIf(v.Flag != nil, "both v.Enum and v.Flag set in %#v", v)
		serSet(v.Enum.Set, indent+1)
	}

	if v.Flag != nil {
		panicIf(v.Enum != nil, "both v.Enum and v.Flag set %#v", v)
		serSet(v.Flag.Set, indent+1)
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

func toInt(s string) (int, bool) {
	n, err := strconv.Atoi(s)
	if err == nil {
		return n, true
	}
	if !strings.HasPrefix(s, "0x") {
		return 0, false
	}
	s = strings.TrimPrefix(s, "0x")
	n2, err := strconv.ParseInt(s, 16, 64)
	if err == nil {
		return int(n2), true
	}
	return 0, false
}

func sortSet(set []*xmldef.Set) {
	nFailed := 0
	var ok bool
	for _, st := range set {
		st.ValueInt, ok = toInt(st.Value)
		if !ok {
			nFailed++
		}
	}
	if nFailed > 1 {
		return
	}
	sort.Slice(set, func(i, j int) bool {
		return set[i].ValueInt < set[j].ValueInt
	})
}

func serSet(set []*xmldef.Set, indent int) {
	sortSet(set)
	maxLen := 0
	for _, s := range set {
		if len(s.Name) > maxLen {
			maxLen = len(s.Name)
		}
	}

	for _, s := range set {
		nSpaces := maxLen - len(s.Name)
		spaces := strings.Repeat(" ", nSpaces)
		indentStr := strings.Repeat("  ", indent)
		outf("%s%s %s%s\n", indentStr, s.Name, spaces, s.Value)
	}
}

func addEnum(args *Args, e *xmldef.Enum) {
	if e == nil {
		return
	}
	args.AddNotEmpty("defaultName", e.DefaultName)
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
		serVariable(v, 0)
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
	serVariables(h.Variable, 0)
}

func serVariables(vars []*xmldef.Variable, indent int) {
	for _, v := range vars {
		serVariable(v, indent)
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

func serModule(m *xmldef.Module) {
	args := NewArgs("dll", escIfNeeded(m.Name))
	args.AddNotEmpty("callingConvention", m.CallingConvention)
	args.AddNotEmpty("errorFunc", m.ErrorFunc)
	args.AddBool("errorIsReturnValue", m.ErrorIsReturnValue)
	args.AddNotEmpty("onlineHelp", m.OnlineHelp)
	args.AddNotEmpty("category", m.Category)
	outf("%s\n", args.String())
	serModuleAlias(m.ModuleAlias)
	serSourceModules(m.SourceModule)
	serCategories(m.Categories)
	serErrorDeocde(m.ErrorDecode)
	serConditions(m.Condition)
	serVariables(m.Variable, 0)
	serApis(m.Api, 0)
}

func serSourceModule(sm *xmldef.SourceModule) {
	panicIf(sm.Name == "", "sm.Name is empty string in %#v", sm)
	args := NewArgs("sourceModule", escIfNeeded(sm.Name))
	args.AddNotEmpty("copy", sm.Copy)
	args.AddNotEmpty("include", sm.Include)
	outf("%s\n", args.String())
	serApis(sm.Api, 0)
}

func serSourceModules(a []*xmldef.SourceModule) {
	for _, sm := range a {
		serSourceModule(sm)
	}
}

func serModuleAlias(a []*xmldef.ModuleAlias) {
	for _, ma := range a {
		args := NewArgs("moduleAlias", escIfNeeded(ma.Name))
		outf("%s\n", args.String())
	}
}

func serErrorDeocde(a []*xmldef.ErrorDecode) {
	for _, ed := range a {
		args := NewArgs("errorDecode")
		args.AddNotEmpty("errorFunc", ed.ErrorFunc)
		args.AddBool("errorIsReturnValue", ed.ErrorIsReturnValue)
		// skip empty errorDecode
		if len(args.a) > 1 {
			outf("%s\n", args.String())
		}
	}
}

func serCategories(categories []*xmldef.Category) {
	for _, c := range categories {
		args := NewArgs("category", escIfNeeded(c.Name))
		args.AddNotEmpty("onlineHelp", c.OnlineHelp)
		outf("%s\n", args.String())
	}
}

func serApis(apis []*xmldef.Api, indent int) {
	for _, api := range apis {
		serApi(api, indent)
	}
}

func serInterface(i *xmldef.Interface) {
	if i == nil {
		return
	}
	args := NewArgs("ingterface", escIfNeeded(i.Name))

	args.AddNotEmpty("base", i.BaseInterface)
	args.AddNotEmpty("id", i.Id)
	args.AddNotEmpty("errorFunc", i.ErrorFunc)
	args.AddNotEmpty("onlineHelp", i.OnlineHelp)
	args.AddNotEmpty("category", i.Category)
	outf("%s\n", args.String())
	serApis(i.Api, 1)
	serVariables(i.Variable, 1)
	outf("\n")
}

func serApi(api *xmldef.Api, indent int) {
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
	args.AddNotEmpty("errorFunc", api.ErrorFunc)
	args.AddBool("errorIsReturnValue", api.ErrorIsReturnValue)
	args.AddBool("varArgs", api.VarArgs)
	args.AddNotEmpty("maxVarArgs", api.MaxVarArgs)
	indentStr := strings.Repeat("  ", indent)
	outf("%s%s\n", indentStr, args.String())
	serSuccess(api.Success, indent+1)
	serReturn(api.Return, indent+1)
	serParams(api.Param, indent+1)
	outf("\n")
}

func serParams(params []*xmldef.Param, indent int) {
	for _, p := range params {
		serParam(p, indent)
	}
}

func serReturn(r *xmldef.Return, indent int) {
	if r == nil {
		return
	}
	args := NewArgs(escIfNeeded(r.Type))
	args.AddNotEmpty("display", r.Display)
	indentStr := strings.Repeat("  ", indent)
	outf("%s%s\n", indentStr, args.String())
}

func serSuccess(s *xmldef.Success, indent int) {
	if s == nil {
		return
	}
	args := NewArgs("success")
	args.AddNotEmpty(s.Return, s.Value)
	args.AddNotEmpty("errorFunc", s.ErrorFunc)
	if len(args.a) == 1 {
		return // skip sucess with no value
	}
	indentStr := strings.Repeat("  ", indent)
	outf("%s%s\n", indentStr, args.String())
}

func toTxt(f *ParsedXmlFile) {
	d := f.Data
	// goes first as it's logically needed first
	serIncludes(d.Include)

	// ignore d.ApiSetSchema
	// ignore d.ErrorLookupModule
	serHeaders(d.Headers)
	// ignore d.HelpUrl
	// d.Include was serialized above
	serInterface(d.Interface)
	for _, m := range d.Module {
		serModule(m)
	}
	// TODO: ensure exclusivity i.e. if Module then no Interface or Headers
}

func skipFile(path string) bool {
	if strings.Contains(path, "Internal") {
		return true
	}
	if strings.Contains(path, "MAPI") {
		return true
	}
	if strings.Contains(path, "MMF") {
		return true
	}
	if strings.Contains(path, "Mozilla") {
		return true
	}
	if strings.Contains(path, "SMI") {
		return true
	}
	if strings.Contains(path, "VSS") {
		return true
	}
	return false
}

//  API/WMI/IWbemShutdown.xml => defs/WMI/IWbemShutdown.txt
func convertFileName(path string) string {
	s := strings.Replace(path, ".xml", ".txt", -1)
	parts := strings.Split(s, "/")
	parts[0] = "defs"
	return filepath.Join(parts...)
}

func createDirForPath(path string) {
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, 0755)
	must(err)
}

func allToTxt() {
	for _, f := range parsedFiles {
		if skipFile(f.Name) {
			continue
		}
		txtPath := convertFileName(f.Name)
		fmt.Printf("%s => %s\n", f.Name, txtPath)
		createDirForPath(txtPath)
		outFile, err := os.Create(txtPath)
		must(err)
		out = outFile
		toTxt(f)
		err = outFile.Close()
		must(err)
	}
}

func main() {
	out = os.Stdout
	timeStart := time.Now()
	parseApiMonitorData()
	fmt.Printf("Parsed %d files in %s\n", len(parsedFiles), time.Since(timeStart))

	allToTxt()
}
