package utils

import (
	"unsafe"
)

// Bytes2Str bytes转string
func Bytes2Str(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// Str2bytes 转换出来的bytes无法修改，只能读取，否则会报指针异常
func Str2bytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}
