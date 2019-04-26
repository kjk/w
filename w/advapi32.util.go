package w

import (
	"errors"
	"fmt"
	"strings"
	"unsafe"
)

func newRegistryError(res uint32, args ...interface{}) error {
	if res == ERROR_SUCCESS {
		return nil
	}
	var msg string
	if len(args) > 0 {
		msg = args[0].(string)
		if len(args) > 1 {
			msg = fmt.Sprintf(msg, args[1:]...)
		}
	}
	if msg != "" {
		msg = strings.TrimSuffix(msg, ".")
		msg += ". "
	}
	sysMsg, _ := FormatSystemMessage(res)
	// TODO: return name of the error constant in addition to error code
	if sysMsg == "" {
		msg += fmt.Sprintf("registry operation failed and returned error code %d", res)
	} else {
		msg += fmt.Sprintf("registry operation failed and returned error code %d (%s)", res, sysMsg)

	}
	return errors.New(msg)
}

// RegOpenKeyEx opens or creates a a registry key
func RegOpenKeyEx(hkey HKEY, subKey string, ulOptions uint32, samDesired uint32) (HKEY, error) {
	u := ToUnicodeShortLived(subKey)
	var hkeyRes HKEY
	res := RegOpenKeyExWSys(
		hkey,
		&u[0],
		ulOptions,
		samDesired,
		&hkeyRes)
	FreeShortLivedUnicode(u)
	return hkeyRes, newRegistryError(res, "RegOpenKeyExW")
}

// RegCreateKeyEx creates a new key or opens existing registry key with KEY_ALL_ACCESS
// returns key, true if created (false if opened) and error
func RegCreateKeyEx(hkey HKEY, subKey string) (HKEY, bool, error) {
	var hkeyRes HKEY
	u := ToUnicodeShortLived(subKey)
	var disp uint32
	res := RegCreateKeyExWSys(
		hkey,
		&u[0],
		0,   // Reserved
		nil, // lpClass
		0,   // dwOptions
		KEY_ALL_ACCESS,
		nil, // lpSecurityAttributes
		&hkeyRes,
		&disp)
	FreeShortLivedUnicode(u)
	didCreate := (disp == REG_CREATED_NEW_KEY)
	return hkeyRes, didCreate, newRegistryError(res, "RecCreateKeyExW")
}

// RegOpenKeyForWrite opens a registry key for writing
func RegOpenKeyForWrite(hkey HKEY, subKey string) (HKEY, error) {
	return RegOpenKeyEx(hkey, subKey, 0, KEY_WRITE)
}

// RegOpenKeyForRead opens a registry key for reading
func RegOpenKeyForRead(hkey HKEY, subKey string) (HKEY, error) {
	return RegOpenKeyEx(hkey, subKey, 0, KEY_READ)
}

// RegCloseKey closes registry key
func RegCloseKey(hkey HKEY) error {
	res := RegCloseKeySys(hkey)
	return newRegistryError(res, "RegCloseKey")
}

// RegWriteString writes a given registry value to a hkey
// (which must opened with RegOpenKeyForWrite)
func RegWriteString(hkey HKEY, subKeyName string, value string) error {
	u := ToUnicodeShortLived(subKeyName)
	su := ToUnicode(value)
	// +2 for terminating 0
	cbData := (len(su) * 2) + 2
	data := (*byte)(unsafe.Pointer(&su[0]))
	res := RegSetValueExWSys(hkey, &u[0], 0, REG_SZ, data, uint32(cbData))
	FreeShortLivedUnicode(u)
	return newRegistryError(res, "RegSetValueExWSys")
}

// RegWriteDWORD writes DWORD registry value
func RegWriteDWORD(hkey HKEY, subKeyName string, value uint32) error {
	u := ToUnicodeShortLived(subKeyName)
	data := (*byte)(unsafe.Pointer(&value))
	cbData := unsafe.Sizeof(value)
	res := RegSetValueExWSys(hkey, &u[0], 0, REG_SZ, data, uint32(cbData))
	FreeShortLivedUnicode(u)
	return newRegistryError(res, "RegSetValueExWSys")
}

// RegDeleteKeyEx deletes registry key
func RegDeleteKeyEx(hkey HKEY, subKey string) error {
	u := ToUnicodeShortLived(subKey)
	res := RegDeleteKeyExWSys(hkey, &u[0], 0, 0)
	FreeShortLivedUnicode(u)
	return newRegistryError(res, "RegDeleteKeyExWSys")
}
