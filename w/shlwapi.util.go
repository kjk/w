package w

// SHDeleteKey deletes a registry subkey and all its descendants.
// https://docs.microsoft.com/en-us/windows/desktop/api/shlwapi/nf-shlwapi-shdeletekeyw
func SHDeleteKey(hkey HKEY, s string) error {
	u := ToUnicodeShortLived(s)
	res := SHDeleteKeyWSys(hkey, &u[0])
	FreeShortLivedUnicode(u)
	return newRegistryError(res)
}
