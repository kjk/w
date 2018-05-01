package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

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

func main() {
	m := loadFiles()
	fmt.Printf("Parsed %d files\n", len(m))
}
