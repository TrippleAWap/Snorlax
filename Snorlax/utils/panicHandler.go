package utils

import (
	"fmt"
	"os"
	"runtime"
	"syscall"
	"unsafe"
)

func PanicHandler() {
	if r := recover(); r != nil {
		// spawn message box with error message
		user32 := syscall.NewLazyDLL("user32.dll")
		MessageBox := user32.NewProc("MessageBoxW")

		stackBuf := make([]byte, 2048)
		stackSize := runtime.Stack(stackBuf, true)

		message := fmt.Sprintf("Panic Occured: %v\n\n%s", r, stackBuf[:stackSize])
		title := "Whoops, something went wrong!"

		titlePtr, _ := syscall.UTF16PtrFromString(title)
		messagePtr, _ := syscall.UTF16PtrFromString(message)

		_, _, err := MessageBox.Call(
			0,
			uintptr(unsafe.Pointer(messagePtr)),
			uintptr(unsafe.Pointer(titlePtr)),
			0x00000010,
		)
		if err != nil && err.Error() != "The operation completed successfully." {
			panic(err)
		}
		os.Exit(1)
	}
}
