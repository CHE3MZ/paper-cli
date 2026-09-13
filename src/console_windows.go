//go:build windows

package src

import (
	"syscall"
	"unsafe"
)

// enableConsoleANSI turns on virtual-terminal processing so ANSI colors
// render on Windows consoles. Best effort: failures just mean plain output.
func enableConsoleANSI() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	setMode := kernel32.NewProc("SetConsoleMode")
	getMode := kernel32.NewProc("GetConsoleMode")
	getHandle := kernel32.NewProc("GetStdHandle")
	const stdOutputHandle = ^uintptr(0) - 10 // (DWORD)-11
	const enableVT = 0x4

	h, _, _ := getHandle.Call(stdOutputHandle)
	if h == 0 {
		return
	}
	var mode uint32
	r, _, _ := getMode.Call(h, uintptr(unsafe.Pointer(&mode)))
	if r == 0 {
		return
	}
	setMode.Call(h, uintptr(mode|enableVT))
}
