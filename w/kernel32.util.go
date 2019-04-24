package w

// GetDriveType returns drive type for a given root dir (like "c:\")
func GetDriveType(rootPathName string) int {
	s := ToUnicodeShortLived(rootPathName)
	res := GetDriveTypeWSys(&s[0])
	FreeShortLivedUnicode(s)
	return int(res)
}
