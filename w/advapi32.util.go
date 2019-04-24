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

// RegOpenKeyForWrite opens a registry key for writing
func RegOpenKeyForWrite(hkey HKEY, subKey string) (HKEY, error) {
	vu := ToUnicodeShortLived(subKey)
	var hKeyRes HKEY
	res := RegOpenKeyExWSys(
		hkey,
		vu,
		0,
		KEY_WRITE,
		&hKeyRes)
	FreeShortLivedUnicode(vu)
	return hKeyRes, newRegistryError(res, "RegOpenKeyExW")
}

// RegWriteString writes a given registry value to a hkey
// (which must opened with RegOpenKeyForWrite)
func RegWriteString(hkey HKEY, subKeyName string, value string) error {
	vu := ToUnicodeShortLived(subKeyName)
	su := ToUnicode(value)
	// includes terminating 0, which su already has
	cbData := len(su) * 2
	data := (*byte)(unsafe.Pointer(&su[0]))
	res := RegSetValueExWSys(hkey, vu, 0, REG_SZ, data, uint32(cbData))
	FreeShortLivedUnicode(vu)
	return newRegistryError(res, "RegSetValueExW")
}

// RegWriteDWORD writes DWORD registry value
func RegWriteDWORD(hkey HKEY, subKeyName string, value uint32) error {
	vu := ToUnicodeShortLived(subKeyName)
	data := (*byte)(unsafe.Pointer(&value))
	cbData := unsafe.Sizeof(value)
	res := RegSetValueExWSys(hkey, vu, 0, REG_SZ, data, uint32(cbData))
	FreeShortLivedUnicode(vu)
	return newRegistryError(res, "RegSetValueExW")
}
