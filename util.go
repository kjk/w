package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"strings"
)

func must(err error) {
	if err != nil {
		panic(err.Error())
	}
}

// FmtArgs formats args as a string. First argument should be format string
// and the rest are arguments to the format
func FmtArgs(args ...interface{}) string {
	if len(args) == 0 {
		return ""
	}
	format := args[0].(string)
	if len(args) == 1 {
		return format
	}
	return fmt.Sprintf(format, args[1:]...)
}

func panicWithMsg(defaultMsg string, args ...interface{}) {
	s := FmtArgs(args...)
	if s == "" {
		s = defaultMsg
	}
	//fmt.Fprintf(os.Stderr, "%s\n", s)
	panic(s)
}

func panicIf(cond bool, args ...interface{}) {
	if !cond {
		return
	}
	panicWithMsg("PanicIf: condition failed", args...)
}

func mustParseBool(s string) bool {
	switch strings.ToLower(s) {
	case "false":
		return false
	case "true":
		return true
	}
	err := fmt.Errorf("'%s' is not a valid bool value", s)
	must(err)
	return false
}

func readZipFile(fi *zip.File) ([]byte, error) {
	r, err := fi.Open()
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	if err != nil {
		r.Close()
		return nil, err
	}
	err = r.Close()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
