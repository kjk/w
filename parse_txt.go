package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Pointer represents a pointer type definition
// pointer HBLUETOOTH_AUTHENTICATION_REGISTRATION* HBLUETOOTH_AUTHENTICATION_REGISTRATION
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

// Defs represents content of a single def file, parsed
type Defs struct {
	// which file was parsed to extract this information, relative to defs directory
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

// Parser for .txt format
type Parser struct {
	Lines    []string
	CurrLine int
}

// RawParsedLine represents a parsed line
type RawParsedLine struct {
	Indent    int
	Strings   []string
	KeyValues []KV
}

// LineType represents type of a parsed line
type LineType int

const (
	// LineEOF means end of file
	LineEOF LineType = iota
	// LineEmpty means empty line
	LineEmpty
	// LineParsed represents RawParsedLine
	LineParsed
)

// TokenType describes type of the token
type TokenType int

const (
	// TokenEOF means end of file
	TokenEOF TokenType = iota
	// TokenString means a single string
	TokenString
	// TokenKV means a key/value pair of strings
	TokenKV
)

func skipSpaces(sp *string) int {
	s := *sp
	nSpaces := 0
	for len(s) > 0 && s[0] == ' ' {
		s = s[1:]
		nSpaces++
	}
	*sp = s
	return nSpaces
}

func parseMaybeQuoted(sp *string) string {
	s := *sp
	tmp := s

	defer func() {
		*sp = s
	}()

	isQuoted := s[0] == '"'
	if isQuoted {
		s = s[1:]
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == ' ' {
			panicIf(isQuoted, "has start quote without end quote in '%s'", tmp)
			s = s[i:]
			return s[:i-1]
		} else if c == '"' {
			panicIf(isQuoted, "found end quote without start in '%s'", tmp)
			s = s[i:]
			return s[:i-1]
		} else if c == '=' {
			s = s[i:]
			return s[:i-1]
		}
	}
	panicIf(isQuoted, "has start quote without end quote in '%s'", tmp)
	return s
}

func nextToken(sp *string) (TokenType, interface{}) {
	s := *sp
	defer func() {
		*sp = s
	}()

	if len(s) == 0 {
		return TokenEOF, nil
	}
	// it's either a single maybe-quoted string or key=value pair where value is maybe quoted
	k := parseMaybeQuoted(&s)
	if len(s) == 0 {
		panicIf(len(k) == 0, "len(k) == 0")
		return TokenString, k
	}
	if s[0] == ' ' {
		skipSpaces(&s)
		return TokenString, k
	}
	panicIf(s[0] != '=', "expected s ('%s') to start with =", s)
	v := parseMaybeQuoted(&s)
	kv := KV{
		Name:  k,
		Value: v,
	}
	if len(s) == 0 {
		panicIf(len(v) == 0, "len(v) == 0")
		return TokenKV, kv
	}
	panicIf(s[0] != ' ', "expected s ('%s') to start with ' '", s)
	skipSpaces(&s)
	return TokenKV, kv
}

func parseLine(s string) *RawParsedLine {
	var res RawParsedLine
	// calc indent from number of spaces at the beginning
	i := skipSpaces(&s)
	panicIf(i%2 == 1, "number of spaces is %d, should be even", i)
	res.Indent = i / 2
	s = s[i:]
	for {
		token, el := nextToken(&s)
		if token == TokenEOF {
			break
		}
		if token == TokenString {
			v := el.(string)
			res.Strings = append(res.Strings, v)
		}
		if token == TokenKV {
			v := el.(KV)
			res.KeyValues = append(res.KeyValues, v)
		}
	}
	return &res
}

// NextLine returns next parsed line
func (p *Parser) NextLine() (LineType, RawParsedLine) {
	var res RawParsedLine
	if p.CurrLine >= len(p.Lines) {
		return LineEOF, res
	}
	l := p.Lines[p.CurrLine]
	p.CurrLine++
	if len(strings.TrimSpace(l)) == 0 {
		return LineEmpty, res
	}
	return LineParsed, res
}

func parseFile(d []byte) *Defs {
	/*
		lines := toLines(d)
		lines = collapseMultipleEmptyLines(lines)
		parser := &Parser{
			Lines: lines,
		}
	*/
	return nil
}

// DefFile represents a conten tof a single def file
type DefFile struct {
	FilePath string // path reltative to defs directory, Unix style (using '/' for path separator)
	Data     []byte
	Defs     *Defs
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
			FilePath: name,
			Data:     d,
			Defs:     defs,
		}
		res[name] = df
		return nil
	}
	filepath.Walk("defs", fn)
	return res
}
