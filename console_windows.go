//go:build windows

package main

import (
	"os"
	"syscall"
)

// 旧来のコンソール (conhost) は既定が CP932 で、UTF-8 の日本語が文字化けする。
// 出力・入力ともコードページを UTF-8 (65001) に切り替える。
func setupConsole() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	kernel32.NewProc("SetConsoleOutputCP").Call(65001)
	kernel32.NewProc("SetConsoleCP").Call(65001)
}

// コンソールに繋がっていれば GetConsoleMode が成功する。
// エクスプローラからのダブルクリック起動がこれに当たる。
func stdinIsTerminal() bool {
	var mode uint32
	return syscall.GetConsoleMode(syscall.Handle(os.Stdin.Fd()), &mode) == nil
}
