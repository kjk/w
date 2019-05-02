// THIS FILE WAS AUTO-GENERATED BY https://github.com/kjk/w/cmd/gengo
package w

import (
	"golang.org/x/sys/windows"
	"syscall"
	"unsafe"
)

var (
	libkernel32 *windows.LazyDLL

	createToolhelp32Snapshot        *windows.LazyProc
	fileTimeToSystemTime            *windows.LazyProc
	formatMessageW                  *windows.LazyProc
	getCurrentThreadId              *windows.LazyProc
	getDiskFreeSpaceExW             *windows.LazyProc
	getDriveTypeW                   *windows.LazyProc
	getFileAttributesExW            *windows.LazyProc
	getLastError                    *windows.LazyProc
	getLogicalDriveStringsW         *windows.LazyProc
	getSystemTimeAsFileTime         *windows.LazyProc
	getTempPathW                    *windows.LazyProc
	getVolumeInformationW           *windows.LazyProc
	heap32First                     *windows.LazyProc
	heap32ListFirst                 *windows.LazyProc
	module32FirstW                  *windows.LazyProc
	module32NextW                   *windows.LazyProc
	multiByteToWideChar             *windows.LazyProc
	process32FirstW                 *windows.LazyProc
	process32NextW                  *windows.LazyProc
	thread32First                   *windows.LazyProc
	thread32Next                    *windows.LazyProc
	toolhelp32ReadProcessMemory     *windows.LazyProc
	tzSpecificLocalTimeToSystemTime *windows.LazyProc
)

func init() {
	libkernel32 = windows.NewLazySystemDLL("kernel32.dll")
	createToolhelp32Snapshot = libkernel32.NewProc("CreateToolhelp32Snapshot")
	fileTimeToSystemTime = libkernel32.NewProc("FileTimeToSystemTime")
	formatMessageW = libkernel32.NewProc("FormatMessageW")
	getCurrentThreadId = libkernel32.NewProc("GetCurrentThreadId")
	getDiskFreeSpaceExW = libkernel32.NewProc("GetDiskFreeSpaceExW")
	getDriveTypeW = libkernel32.NewProc("GetDriveTypeW")
	getFileAttributesExW = libkernel32.NewProc("GetFileAttributesExW")
	getLastError = libkernel32.NewProc("GetLastError")
	getLogicalDriveStringsW = libkernel32.NewProc("GetLogicalDriveStringsW")
	getSystemTimeAsFileTime = libkernel32.NewProc("GetSystemTimeAsFileTime")
	getTempPathW = libkernel32.NewProc("GetTempPathW")
	getVolumeInformationW = libkernel32.NewProc("GetVolumeInformationW")
	heap32First = libkernel32.NewProc("Heap32First")
	heap32ListFirst = libkernel32.NewProc("Heap32ListFirst")
	module32FirstW = libkernel32.NewProc("Module32FirstW")
	module32NextW = libkernel32.NewProc("Module32NextW")
	multiByteToWideChar = libkernel32.NewProc("MultiByteToWideChar")
	process32FirstW = libkernel32.NewProc("Process32FirstW")
	process32NextW = libkernel32.NewProc("Process32NextW")
	thread32First = libkernel32.NewProc("Thread32First")
	thread32Next = libkernel32.NewProc("Thread32Next")
	toolhelp32ReadProcessMemory = libkernel32.NewProc("Toolhelp32ReadProcessMemory")
	tzSpecificLocalTimeToSystemTime = libkernel32.NewProc("TzSpecificLocalTimeToSystemTime")
}

const (
	TH32CS_INHERIT      = 0x80000000
	TH32CS_SNAPALL      = 0x0000000f
	TH32CS_SNAPHEAPLIST = 0x00000001
	TH32CS_SNAPMODULE   = 0x00000008
	TH32CS_SNAPMODULE32 = 0x00000010
	TH32CS_SNAPPROCESS  = 0x00000002
	TH32CS_SNAPTHREAD   = 0x00000004
)

const (
	FORMAT_MESSAGE_ALLOCATE_BUFFER = 0x00000100
	FORMAT_MESSAGE_IGNORE_INSERTS  = 0x00000200
	FORMAT_MESSAGE_FROM_STRING     = 0x00000400
	FORMAT_MESSAGE_FROM_HMODULE    = 0x00000800
	FORMAT_MESSAGE_FROM_SYSTEM     = 0x00001000
	FORMAT_MESSAGE_ARGUMENT_ARRAY  = 0x00002000
	FORMAT_MESSAGE_MAX_WIDTH_MASK  = 0x000000FF
)

const (
	DRIVE_UNKNOWN     = 0
	DRIVE_NO_ROOT_DIR = 1
	DRIVE_REMOVABLE   = 2
	DRIVE_FIXED       = 3
	DRIVE_REMOTE      = 4
	DRIVE_CDROM       = 5
	DRIVE_RAMDISK     = 6
)

const (
	FILE_CASE_SENSITIVE_SEARCH        = 0x00000001
	FILE_CASE_PRESERVED_NAMES         = 0x00000002
	FILE_UNICODE_ON_DISK              = 0x00000004
	FILE_PERSISTENT_ACLS              = 0x00000008
	FILE_FILE_COMPRESSION             = 0x00000010
	FILE_VOLUME_QUOTAS                = 0x00000020
	FILE_SUPPORTS_SPARSE_FILES        = 0x00000040
	FILE_SUPPORTS_REPARSE_POINTS      = 0x00000080
	FILE_SUPPORTS_REMOTE_STORAGE      = 0x00000100
	FILE_VOLUME_IS_COMPRESSED         = 0x00008000
	FILE_SUPPORTS_OBJECT_IDS          = 0x00010000
	FILE_SUPPORTS_ENCRYPTION          = 0x00020000
	FILE_NAMED_STREAMS                = 0x00040000
	FILE_READ_ONLY_VOLUME             = 0x00080000
	FILE_SEQUENTIAL_WRITE_ONCE        = 0x00100000
	FILE_SUPPORTS_TRANSACTIONS        = 0x00200000
	FILE_SUPPORTS_HARD_LINKS          = 0x00400000
	FILE_SUPPORTS_EXTENDED_ATTRIBUTES = 0x00800000
	FILE_SUPPORTS_OPEN_BY_FILE_ID     = 0x01000000
	FILE_SUPPORTS_USN_JOURNAL         = 0x02000000
	FILE_SUPPORTS_INTEGRITY_STREAMS   = 0x04000000
)

const (
	LF32_FIXED    = 0x00000001
	LF32_FREE     = 0x00000002
	LF32_MOVEABLE = 0x00000004
)

type HEAPENTRY32 struct {
	DwSize        int32
	HHandle       uintptr
	DwAddress     uintptr
	DwBlockSize   int32
	DwFlags       uint32
	DwLockCount   uint32
	DwResvd       uint32
	Th32ProcessID uint32
	Th32HeapID    uintptr
}

const (
	HF32_DEFAULT = 1
	HF32_SHARED  = 2
)

type HEAPLIST32 struct {
	DwSize        int32
	Th32ProcessID uint32
	Th32HeapID    uintptr
	DwFlags       uint32
}

type MODULEENTRY32 struct {
	DwSize        uint32
	Th32ModuleID  uint32
	Th32ProcessID uint32
	GlblcntUsage  uint32
	ProccntUsage  uint32
	ModBaseAddr   *uint8
	ModBaseSize   uint32
	HModule       HANDLE
	SzModule      [256]WCHAR
	SzExePath     [260]WCHAR
}

const (
	MB_PRECOMPOSED       = 0x00000001
	MB_COMPOSITE         = 0x00000002
	MB_USEGLYPHCHARS     = 0x00000004
	MB_ERR_INVALID_CHARS = 0x00000008
)

type PROCESSENTRY32 struct {
	DwSize              uint32
	CntUsage            uint32
	Th32ProcessID       uint32
	Th32DefaultHeapID   uintptr
	Th32ModuleID        uint32
	CntThreads          uint32
	Th32ParentProcessID uint32
	PcPriClassBase      int32
	DwFlags             uint32
	SzExeFile           [260]WCHAR
}

type THREADENTRY32 struct {
	DwSize             uint32
	CntUsage           uint32
	Th32ThreadID       uint32
	Th32OwnerProcessID uint32
	TpBasePri          int32
	TpDeltaPri         int32
	DwFlags            uint32
}

func CreateToolhelp32SnapshotSys(dwFlags uint32, th32ProcessID uint32) HANDLE {
	ret, _, _ := syscall.Syscall(createToolhelp32Snapshot.Addr(), 2,
		uintptr(dwFlags),
		uintptr(th32ProcessID),
		0,
	)
	return HANDLE(ret)
}

func FileTimeToSystemTimeSys(lpFileTime *FILETIME, lpSystemTime *SYSTEMTIME) int32 {
	ret, _, _ := syscall.Syscall(fileTimeToSystemTime.Addr(), 2,
		uintptr(unsafe.Pointer(lpFileTime)),
		uintptr(unsafe.Pointer(lpSystemTime)),
		0,
	)
	return int32(ret)
}

func FormatMessageWSys(dwFlags uint32, lpSource unsafe.Pointer, dwMessageId uint32, dwLanguageId uint32, lpBuffer *WCHAR, nSize uint32, Arguments *unsafe.Pointer) uint32 {
	ret, _, _ := syscall.Syscall9(formatMessageW.Addr(), 7,
		uintptr(dwFlags),
		uintptr(lpSource),
		uintptr(dwMessageId),
		uintptr(dwLanguageId),
		uintptr(unsafe.Pointer(lpBuffer)),
		uintptr(nSize),
		uintptr(unsafe.Pointer(Arguments)),
		0,
		0,
	)
	return uint32(ret)
}

func GetCurrentThreadIdSys() uint32 {
	ret, _, _ := syscall.Syscall(getCurrentThreadId.Addr(), 0,
		0,
		0,
		0,
	)
	return uint32(ret)
}

func GetDiskFreeSpaceExWSys(lpDirectoryName *uint16, lpFreeBytesAvailable *ULARGE_INTEGER, lpTotalNumberOfBytes *ULARGE_INTEGER, lpTotalNumberOfFreeBytes *ULARGE_INTEGER) int32 {
	ret, _, _ := syscall.Syscall6(getDiskFreeSpaceExW.Addr(), 4,
		uintptr(unsafe.Pointer(lpDirectoryName)),
		uintptr(unsafe.Pointer(lpFreeBytesAvailable)),
		uintptr(unsafe.Pointer(lpTotalNumberOfBytes)),
		uintptr(unsafe.Pointer(lpTotalNumberOfFreeBytes)),
		0,
		0,
	)
	return int32(ret)
}

func GetDriveTypeWSys(lpRootPathName *uint16) uint32 {
	ret, _, _ := syscall.Syscall(getDriveTypeW.Addr(), 1,
		uintptr(unsafe.Pointer(lpRootPathName)),
		0,
		0,
	)
	return uint32(ret)
}

func GetFileAttributesExWSys(lpFileName *uint16, fInfoLevelId uint32, lpFileInformation unsafe.Pointer) int32 {
	ret, _, _ := syscall.Syscall(getFileAttributesExW.Addr(), 3,
		uintptr(unsafe.Pointer(lpFileName)),
		uintptr(fInfoLevelId),
		uintptr(lpFileInformation),
	)
	return int32(ret)
}

func GetLastErrorSys() uint32 {
	ret, _, _ := syscall.Syscall(getLastError.Addr(), 0,
		0,
		0,
		0,
	)
	return uint32(ret)
}

func GetLogicalDriveStringsWSys(nBufferLength uint32, lpBuffer *WCHAR) uint32 {
	ret, _, _ := syscall.Syscall(getLogicalDriveStringsW.Addr(), 2,
		uintptr(nBufferLength),
		uintptr(unsafe.Pointer(lpBuffer)),
		0,
	)
	return uint32(ret)
}

func GetSystemTimeAsFileTimeSys(lpSystemTimeAsFileTime *FILETIME) {
	_, _, _ = syscall.Syscall(getSystemTimeAsFileTime.Addr(), 1,
		uintptr(unsafe.Pointer(lpSystemTimeAsFileTime)),
		0,
		0,
	)
}

func GetTempPathWSys(nBufferLength uint32, lpBuffer *WCHAR) uint32 {
	ret, _, _ := syscall.Syscall(getTempPathW.Addr(), 2,
		uintptr(nBufferLength),
		uintptr(unsafe.Pointer(lpBuffer)),
		0,
	)
	return uint32(ret)
}

func GetVolumeInformationWSys(lpRootPathName *uint16, lpVolumeNameBuffer *WCHAR, nVolumeNameSize uint32, lpVolumeSerialNumber *uint32, lpMaximumComponentLength *uint32, lpFileSystemFlags *uint32, lpFileSystemNameBuffer *WCHAR, nFileSystemNameSize uint32) int32 {
	ret, _, _ := syscall.Syscall9(getVolumeInformationW.Addr(), 8,
		uintptr(unsafe.Pointer(lpRootPathName)),
		uintptr(unsafe.Pointer(lpVolumeNameBuffer)),
		uintptr(nVolumeNameSize),
		uintptr(unsafe.Pointer(lpVolumeSerialNumber)),
		uintptr(unsafe.Pointer(lpMaximumComponentLength)),
		uintptr(unsafe.Pointer(lpFileSystemFlags)),
		uintptr(unsafe.Pointer(lpFileSystemNameBuffer)),
		uintptr(nFileSystemNameSize),
		0,
	)
	return int32(ret)
}

func Heap32FirstSys(lphe *HEAPENTRY32, th32ProcessID uint32, th32HeapID uintptr) int32 {
	ret, _, _ := syscall.Syscall(heap32First.Addr(), 3,
		uintptr(unsafe.Pointer(lphe)),
		uintptr(th32ProcessID),
		uintptr(th32HeapID),
	)
	return int32(ret)
}

func Heap32ListFirstSys(hSnapshot uintptr, lphl *HEAPLIST32) int32 {
	ret, _, _ := syscall.Syscall(heap32ListFirst.Addr(), 2,
		uintptr(hSnapshot),
		uintptr(unsafe.Pointer(lphl)),
		0,
	)
	return int32(ret)
}

func Module32FirstWSys(hSnapshot uintptr, lpme *MODULEENTRY32) int32 {
	ret, _, _ := syscall.Syscall(module32FirstW.Addr(), 2,
		uintptr(hSnapshot),
		uintptr(unsafe.Pointer(lpme)),
		0,
	)
	return int32(ret)
}

func Module32NextWSys(hSnapshot uintptr, lpme *MODULEENTRY32) int32 {
	ret, _, _ := syscall.Syscall(module32NextW.Addr(), 2,
		uintptr(hSnapshot),
		uintptr(unsafe.Pointer(lpme)),
		0,
	)
	return int32(ret)
}

func MultiByteToWideCharSys(CodePage uint32, dwFlags uint32, lpMultiByteStr *byte, cbMultiByte int32, lpWideCharStr LPWSTR, cchWideChar int32) int32 {
	ret, _, _ := syscall.Syscall6(multiByteToWideChar.Addr(), 6,
		uintptr(CodePage),
		uintptr(dwFlags),
		uintptr(unsafe.Pointer(lpMultiByteStr)),
		uintptr(cbMultiByte),
		uintptr(unsafe.Pointer(lpWideCharStr)),
		uintptr(cchWideChar),
	)
	return int32(ret)
}

func Process32FirstWSys(hSnapshot uintptr, lppe *PROCESSENTRY32) int32 {
	ret, _, _ := syscall.Syscall(process32FirstW.Addr(), 2,
		uintptr(hSnapshot),
		uintptr(unsafe.Pointer(lppe)),
		0,
	)
	return int32(ret)
}

func Process32NextWSys(hSnapshot uintptr, lppe *PROCESSENTRY32) int32 {
	ret, _, _ := syscall.Syscall(process32NextW.Addr(), 2,
		uintptr(hSnapshot),
		uintptr(unsafe.Pointer(lppe)),
		0,
	)
	return int32(ret)
}

func Thread32FirstSys(hSnapshot uintptr, lpte *THREADENTRY32) int32 {
	ret, _, _ := syscall.Syscall(thread32First.Addr(), 2,
		uintptr(hSnapshot),
		uintptr(unsafe.Pointer(lpte)),
		0,
	)
	return int32(ret)
}

func Thread32NextSys(hSnapshot uintptr, lpte *THREADENTRY32) int32 {
	ret, _, _ := syscall.Syscall(thread32Next.Addr(), 2,
		uintptr(hSnapshot),
		uintptr(unsafe.Pointer(lpte)),
		0,
	)
	return int32(ret)
}

func Toolhelp32ReadProcessMemorySys(th32ProcessID uint32, lpBaseAddress unsafe.Pointer, lpBuffer unsafe.Pointer, cbRead int32, lpNumberOfBytesRead int32) int32 {
	ret, _, _ := syscall.Syscall6(toolhelp32ReadProcessMemory.Addr(), 5,
		uintptr(th32ProcessID),
		uintptr(lpBaseAddress),
		uintptr(lpBuffer),
		uintptr(cbRead),
		uintptr(lpNumberOfBytesRead),
		0,
	)
	return int32(ret)
}

func TzSpecificLocalTimeToSystemTimeSys(lpTimeZoneInformation *TIME_ZONE_INFORMATION, lpLocalTime *SYSTEMTIME, lpUniversalTime *SYSTEMTIME) int32 {
	ret, _, _ := syscall.Syscall(tzSpecificLocalTimeToSystemTime.Addr(), 3,
		uintptr(unsafe.Pointer(lpTimeZoneInformation)),
		uintptr(unsafe.Pointer(lpLocalTime)),
		uintptr(unsafe.Pointer(lpUniversalTime)),
	)
	return int32(ret)
}
