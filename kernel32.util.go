package w

import (
	"errors"
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

// FormatSystemMessage returns system error message for a given msgID (error code)
func FormatSystemMessage(msgID uint32) (string, error) {
	// TODO: allocate buffer with FORMAT_MESSAGE_ALLOCATE_BUFFER
	var buf [8 * 1024]uint16
	res := FormatMessageWSys(
		FORMAT_MESSAGE_FROM_SYSTEM,
		nil,
		msgID,
		0, // dwLanguageId
		&buf[0],
		uint32(len(buf)),
		nil)
	if res == 0 {
		return "", errors.New("FormatMessageW failed")
	}
	s := FromUTF16(buf[:len(buf)])
	return s, nil
}

// GetDriveType returns drive type for a given root dir (like "c:\")
func GetDriveType(rootPathName string) int {
	s := ToUTF16ShortLived(rootPathName)
	res := GetDriveTypeWSys(&s[0])
	FreeShortLivedUTF16(s)
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
