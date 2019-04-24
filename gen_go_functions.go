package main

func (g *goGenerator) genFunctions() {
	//g.addFunction("CreateWindowExW")

	functions := []string{
		"MultiByteToWideChar",

		"CoInitialize",
		"CoUninitialize",
		"CoGetClassObject",
		"CoCreateInstance",
		"SHGetFolderPathW",
		"SHGetFolderLocation",

		"RegOpenKeyExW",
		"RegSetValueExW",
		"RegCloseKey",
		"RegDeleteKeyExW",
		"RegSetKeySecurity",
		"RegCreateKeyExW",
		"SHSetValueW",
		"SHGetValueW",
		"SHDeleteValueW",
		"SHDeleteKeyW",
		"SHQueryInfoKeyW",
		"SHQueryValueExW",

		"ImpersonateSelf",
		"InitializeAcl",
		"InitializeSecurityDescriptor",
		"InitializeSid",
		"SetSecurityDescriptorDacl",

		"GetTempPathW",
		"GetVolumeInformationW",
		"GetDriveTypeW",
		"GetLogicalDriveStringsW",
		"GetLastError",
		"FormatMessageW",
		"GetDiskFreeSpaceExW",

		"FileTimeToSystemTime",
		"TzSpecificLocalTimeToSystemTime",
		"GetSystemTimeAsFileTime",

		"MonitorFromRect",
		"GetMonitorInfoW",
		"GetSystemMetrics",
		"SystemParametersInfoW",
		"GetDesktopWindow",
		"FindWindowW",
		"UpdateWindow",

		"CreateToolhelp32Snapshot",
		"Heap32First",
		"Heap32ListFirst",
		"Module32FirstW",
		"Module32NextW",
		"Process32FirstW",
		"Process32NextW",
		"Thread32First",
		"Thread32Next",
		"Toolhelp32ReadProcessMemory",
		"GetFileAttributesExW",
	}

	for _, f := range functions {
		g.addFunction(f)
	}
}

func (g *goGenerator) genInterfaces() {
	interfaces := []string{
		"IClassFactory",
		"IShellLinkW",
		"IPersistFile",
	}
	for _, i := range interfaces {
		g.addInterface(i)
	}
}
