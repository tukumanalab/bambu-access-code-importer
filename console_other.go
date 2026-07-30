//go:build !windows

package main

import "os"

// Windows 以外は端末が UTF-8 なので何もしない。
func setupConsole() {}

func stdinIsTerminal() bool {
	info, err := os.Stdin.Stat()
	if err != nil || info.Mode()&os.ModeCharDevice == 0 {
		return false
	}
	// /dev/null もキャラクタデバイスなので、端末と区別する。
	if null, err := os.Stat(os.DevNull); err == nil && os.SameFile(info, null) {
		return false
	}
	return true
}
