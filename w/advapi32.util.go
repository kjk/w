package w

import (
	"fmt"
	"unsafe"
)

func newRegistryError(res uint32, args ...interface{}) error {
	if res == ERROR_SUCCESS {
		return nil
	}
	// TODO: use args for error message
	// TODO: return name of the error constant in addition to error code
	return fmt.Errorf("registry operation failed with %d", res)
}

// RegOpenKeyEx opens or creates a a registry key
func RegOpenKeyEx(hkey HKEY, subKey string, ulOptions uint32, samDesired uint32) (HKEY, error) {
	vu := ToUnicodeShortLived(subKey)
	var hKeyRes HKEY
	res := RegOpenKeyExWSys(
		hkey,
		&vu[0],
		ulOptions,
		samDesired,
		&hKeyRes)
	FreeShortLivedUnicode(vu)
	return hKeyRes, newRegistryError(res, "RegOpenKeyExW")

}

// RegOpenKeyForWrite opens a registry key for writing
func RegOpenKeyForWrite(hkey HKEY, subKey string) (HKEY, error) {
	return RegOpenKeyEx(hkey, subKey, 0, KEY_WRITE)
}

// RegWriteString writes a given registry value to a hkey
// (which must opened with RegOpenKeyForWrite)
func RegWriteString(hkey HKEY, subKeyName string, value string) error {
	vu := ToUnicodeShortLived(subKeyName)
	su := ToUnicode(value)
	// +2 for terminating 0
	cbData := (len(su) * 2) + 2
	data := (*byte)(unsafe.Pointer(&su[0]))
	res := RegSetValueExWSys(hkey, &vu[0], 0, REG_SZ, data, uint32(cbData))
	FreeShortLivedUnicode(vu)
	return newRegistryError(res, "RegSetValueExWSys")
}

// RegWriteDWORD writes DWORD registry value
func RegWriteDWORD(hkey HKEY, subKeyName string, value uint32) error {
	vu := ToUnicodeShortLived(subKeyName)
	data := (*byte)(unsafe.Pointer(&value))
	cbData := unsafe.Sizeof(value)
	res := RegSetValueExWSys(hkey, &vu[0], 0, REG_SZ, data, uint32(cbData))
	FreeShortLivedUnicode(vu)
	return newRegistryError(res, "RegSetValueExWSys")
}

// RegDeleteKeyEx deletes registry key
func RegDeleteKeyEx(hkey HKEY, subKey string, samDesired uint32) error {
	vu := ToUnicodeShortLived(subKey)
	res := RegDeleteKeyExWSys(hkey, &vu[0], samDesired, 0)
	FreeShortLivedUnicode(vu)
	return newRegistryError(res, "RegDeleteKeyExWSys")
}
