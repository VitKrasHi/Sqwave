//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

var (
	user32       = syscall.NewLazyDLL("user32.dll")
	procMessageB = user32.NewProc("MessageBoxW")
)

// mbIconError — иконка «ошибка» в MessageBox.
const mbIconError = 0x00000010

// showErrorDialog показывает нативное окно с текстом ошибки.
// Нужно потому, что GUI-сборка не имеет консоли.
func showErrorDialog(title, text string) {
	t, err := syscall.UTF16PtrFromString(title)
	if err != nil {
		return
	}
	m, err := syscall.UTF16PtrFromString(text)
	if err != nil {
		return
	}
	_, _, _ = procMessageB.Call(0,
		uintptr(unsafe.Pointer(m)),
		uintptr(unsafe.Pointer(t)),
		mbIconError)
}
