package update

import (
	"syscall"
	"unsafe"
)

// SystemLocale returns the user's UI locale name (e.g. "it-IT").
func SystemLocale() string {
	p, err := syscall.GetProcAddress(mustKernel32(), "GetUserDefaultLocaleName")
	if err != nil {
		return ""
	}
	buf := make([]uint16, 85)
	r, _, _ := syscall.SyscallN(p, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	if r == 0 {
		return ""
	}
	return syscall.UTF16ToString(buf)
}

func mustKernel32() syscall.Handle {
	h, _ := syscall.LoadLibrary("kernel32.dll")
	return h
}
