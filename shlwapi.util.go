package w

// SHDeleteKey deletes a registry subkey and all its descendants.
// https://docs.microsoft.com/en-us/windows/desktop/api/shlwapi/nf-shlwapi-shdeletekeyw
func SHDeleteKey(hkey HKEY, s string) error {
	u := ToUTF16ShortLived(s)
	res := SHDeleteKeyWSys(hkey, &u[0])
	FreeShortLivedUTF16(u)
	return newRegistryError(res)
}
