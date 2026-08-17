//go:build windows

package main

import (
	"syscall"
	"unicode/utf16"
	"unsafe"
)

// FOLDERID_RoamingAppData ({3EB685DB-65F9-4CF6-A03A-E3EF65729F3D})。
var folderIDRoamingAppData = guid{
	0x3EB685DB, 0x65F9, 0x4CF6,
	[8]byte{0xA0, 0x3A, 0xE3, 0xEF, 0x65, 0x72, 0x9F, 0x3D},
}

type guid struct {
	data1 uint32
	data2 uint16
	data3 uint16
	data4 [8]byte
}

var (
	shell32                  = syscall.NewLazyDLL("shell32.dll")
	procSHGetKnownFolderPath = shell32.NewProc("SHGetKnownFolderPath")
	ole32                    = syscall.NewLazyDLL("ole32.dll")
	procCoTaskMemFree        = ole32.NewProc("CoTaskMemFree")
)

// Roaming AppData の実体をシェルに聞く。%APPDATA% は起動の仕方によって
// 古い値のまま引き継がれることがあり (エクスプローラからの起動で、別名だった
// 頃のプロファイルを指していた事例があった)、それを見ていると Bambu Studio が
// 読まない conf を書き換えてしまう。Bambu Studio (wxWidgets) 自身もシェルに
// 聞いているので、こちらに合わせる。
func roamingAppDataDir() string {
	var p *uint16
	r, _, _ := procSHGetKnownFolderPath.Call(
		uintptr(unsafe.Pointer(&folderIDRoamingAppData)),
		0, 0,
		uintptr(unsafe.Pointer(&p)),
	)
	if r != 0 || p == nil {
		return ""
	}
	defer procCoTaskMemFree.Call(uintptr(unsafe.Pointer(p)))

	var buf []uint16
	for q := p; *q != 0; q = (*uint16)(unsafe.Add(unsafe.Pointer(q), 2)) {
		buf = append(buf, *q)
	}
	return string(utf16.Decode(buf))
}
