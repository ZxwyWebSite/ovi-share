package util

import _ "unsafe"

// runtime.bytealg_MakeNoZero

// MakeNoZero 生成长度和容量为 n 的切片，但不将字节清零。调用者有责任确保未初始化的字节不会泄露给最终用户。
//
//go:linkname MakeNoZero internal/bytealg.MakeNoZero
func MakeNoZero(int) []byte
