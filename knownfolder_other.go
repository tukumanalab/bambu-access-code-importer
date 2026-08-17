//go:build !windows

package main

// Windows 以外にシェルの既知フォルダはない。os.UserConfigDir() と
// ホーム由来の候補だけを使う。
func roamingAppDataDir() string { return "" }
