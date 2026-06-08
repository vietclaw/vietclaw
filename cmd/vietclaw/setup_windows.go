//go:build windows

package main

import (
	"os"
	"golang.org/x/sys/windows"
)

func setTerminalRaw() (func(), error) {
	handle := windows.Handle(os.Stdin.Fd())
	var oldMode uint32
	if err := windows.GetConsoleMode(handle, &oldMode); err != nil {
		return func() {}, err
	}
	
	// Disable line input, echo, and processed input. Enable virtual terminal input if possible.
	rawMode := oldMode &^ windows.ENABLE_LINE_INPUT &^ windows.ENABLE_ECHO_INPUT &^ windows.ENABLE_PROCESSED_INPUT
	rawMode |= windows.ENABLE_VIRTUAL_TERMINAL_INPUT
	
	_ = windows.SetConsoleMode(handle, rawMode)
	
	cleanup := func() {
		_ = windows.SetConsoleMode(handle, oldMode)
	}
	return cleanup, nil
}
