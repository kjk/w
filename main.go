package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Pointer represents a pointer type definition
// pointer HBLUETOOTH_AUTHENTICATION_REGISTRATION* HBLUETOOTH_AUTHENTICATION_REGISTRATION
//
//
type Pointer struct {
	Name string
	Type string
}

// Alias represents a type alias definition
// alias PFN_BLUETOOTH_ENUM_ATTRIBUTES_CALLBACK LPVOID
type Alias struct {
	Name string
	Type string
}

// Integer represents an integer type definition
// integer DWORD size=4 unsigned
type Integer struct {
	Name       string
	Size       int
	IsUnsigned bool
}

// Array represents an array definition
// array "WCHAR [BLUETOOTH_MAX_NAME_SIZE]" base=WCHAR count=248
type Array struct {
	Name  string
	Base  string
	Count int
}

// KV is key/value pair (used for Enum and Flag)
type KV struct {
	Name  string
	Value string
}

// Enum represents an enumeration definition.
// Enum is a set of distinct values
// enum [DD_HRESULT] HRESULT display=HRESULT reset
type Enum struct {
	Name    string
	Base    string
	Display string
	Reset   bool
	Values  []KV
}

// Flag represents a flag definition.
// Flag is like Enum but values are meant to be OR'ed together
// flag [DDSCAPS2_FLAGS] DWORD display=DWORD
type Flag struct {
	Name   string
	Values []KV
}

// Field represents a struct field
type Field struct {
}

// Struct represents a struct definition.
// struct [SDP_ELEMENT_DATA_u_s1] display=struct
type Struct struct {
	Name   string
	Fields []Field
}

// Union represents a union definition.
// union [BLUETOOTH_AUTHENTICATE_RESPONSE_u] display=union
type Union struct {
	Name    string
	Display string
}

// Param represents an argument to function.
type Param struct {
	Name string
}

// Func represents a function definition.
type Func struct {
	Name         string
	Return       string
	Params       []Param
	Ordinal      int
	OrdinalA     int
	OrdinaW      int
	BothCharsets bool
}

// Interface represents an interface definition.
// ingterface IDWriteBitmapRenderTarget base=IUnknown id={5e5a32a3-8dff-4773-9ff6-0696eab77267} errorFunc=HRESULT onlineHelp=MSDN category="Graphics and Gaming/DirectX Graphics and Gaming/DirectWrite"
type Interface struct {
	Name      string
	Base      string
	ID        string
	ErrorFunc string
	Functions []Func
}

// Module represents a single .dll file
type Module struct {
	Functions []Func
}

// Defs represents content of a single def file, parsed
type Defs struct {
	// which file was parsed to extract this information
	FilePath string
	Includes []string
	IsHeader bool
	// list of interfaces, only valid if IsHeader is true
	InterfaceDecls []string

	Module *Module

	Aliases    []Alias
	Pointers   []Pointer
	Integers   []Integer
	Enums      []Enum
	Flags      []Flag
	Interfaces []Interface
}

// DefFile represents a conten tof a single def file
type DefFile struct {
	FilePath string // path reltative to defs directory, Unix style (using '/' for path separator)
	Data     []byte
	Defs     *Defs
}

func parseFile(d []byte) *Defs {
	return nil
}

// loadFiles loads all definition files and returns a map of filename
func loadFiles() map[string]*DefFile {
	res := map[string]*DefFile{}
	fn := func(path string, info os.FileInfo, err error) error {
		must(err)
		if info.IsDir() {
			return nil
		}
		unixPath := normalizePath(path)
		// remove first level dir defs from name
		parts := strings.Split(unixPath, "/")
		name := strings.Join(parts[1:], "/")
		fmt.Printf("path: %s, name: %s\n", path, name)
		d := readFileMust(path)
		defs := parseFile(d)
		df := &DefFile{
			Path: name,
			Data: d,
			Defs: defs,
		}
		res[name] = df
		return nil
	}
	filepath.Walk("defs", fn)
	return res
}

func main() {
	m := loadFiles()
	fmt.Printf("Parsed %d files\n", len(m))
}
