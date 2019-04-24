package w

// CoInitialize initializes COM
func CoInitialize() error {
	hr := CoInitializeSys(nil)
	return errorFromHRESULT("CoInitialize", hr)
}
