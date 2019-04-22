package w

// GetKnownFolderPath returns location of known folder
// like CSIDL_APPDATA
func GetKnownFolderPath(nFolder int) (string, error) {
	var buf [MAX_PATH + 1]WCHAR
	hr := SHGetFolderPathW(0, int32(nFolder), 0, SHGFP_TYPE_CURRENT, &buf[0])
	if FAILED(hr) {
		return "", errorFromHRESULT("SHGetFolderPathW", hr)
	}
	return FromUnicode(buf[:]), nil
}
