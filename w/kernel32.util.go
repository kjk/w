package w

import (
	"fmt"
	"syscall"
)

func newGetLastError() error {
	lastError := GetLastErrorSys()
	if lastError == ERROR_SUCCESS {
		return nil
	}
	s, _ := FormatSystemMessage(lastError)
	if s == "" {
		return fmt.Errorf("GetLastError(): %d", lastError)
	}
	return fmt.Errorf("GetLastError(): %d (%s)", lastError, s)
}

// GetDriveType returns drive type for a given root dir (like "c:\")
func GetDriveType(rootPathName string) int {
	s := ToUnicodeShortLived(rootPathName)
	res := GetDriveTypeWSys(&s[0])
	FreeShortLivedUnicode(s)
	return int(res)
}

// converts multiple, 0-terminated utf16 strings into an array
// of utf8 strings
func utf16StringsToList(s []uint16) []string {
	var res []string
	for i := 0; i < len(s)-2; {
		str := syscall.UTF16ToString(s[i:])
		res = append(res, str)
		i += len(str) + 1
	}
	return res
}

// GetLogicalDriveStrings returns a list of logical drives
func GetLogicalDriveStrings() ([]string, error) {
	// find out the size of needed buffer
	n := GetLogicalDriveStringsWSys(0, nil)
	if n == 0 {
		return nil, newGetLastError()
	}
	buf := make([]uint16, n+1, n+1)
	n = GetLogicalDriveStringsWSys(n, &buf[0])
	if n == 0 {
		return nil, newGetLastError()
	}
	return utf16StringsToList(buf), nil
}
