//go:build windows

// One Blackjack at a time: a second launch brings the open window forward.
// (c) 2026 Ironite
package main

import (
	"syscall"
	"unsafe"
)

var (
	k32         = syscall.NewLazyDLL("kernel32.dll")
	u32         = syscall.NewLazyDLL("user32.dll")
	createMutex = k32.NewProc("CreateMutexW")
	findWindow  = u32.NewProc("FindWindowW")
	showWindow  = u32.NewProc("ShowWindow")
	setForegrnd = u32.NewProc("SetForegroundWindow")
	isIconic    = u32.NewProc("IsIconic")
)

const errAlreadyExists = 183

func singleInstance() bool {
	name, _ := syscall.UTF16PtrFromString(`Local\IroniteBlackjack`)
	_, _, err := createMutex.Call(0, 0, uintptr(unsafe.Pointer(name)))
	if errno, ok := err.(syscall.Errno); ok && errno == errAlreadyExists {
		t, _ := syscall.UTF16PtrFromString(title)
		if h, _, _ := findWindow.Call(0, uintptr(unsafe.Pointer(t))); h != 0 {
			if ic, _, _ := isIconic.Call(h); ic != 0 {
				showWindow.Call(h, 9) // SW_RESTORE
			}
			setForegrnd.Call(h)
		}
		return false
	}
	return true
}
