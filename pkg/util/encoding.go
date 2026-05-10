package util

import (
	"encoding/base64"
	"encoding/hex"
)

//# base64

// Base64 编码
func Base64Encode(enc *base64.Encoding, src []byte) []byte {
	dst := MakeNoZero(enc.EncodedLen(len(src)))
	enc.Encode(dst, src)
	return dst
}

// Base64 解码
func Base64Decode(enc *base64.Encoding, src []byte) ([]byte, error) {
	dst := MakeNoZero(enc.DecodedLen(len(src)))
	n, err := enc.Decode(dst, src)
	return dst[:n], err
}

//# hex

// Hex 编码
func HexEncode(src []byte) []byte {
	dst := MakeNoZero(hex.EncodedLen(len(src)))
	hex.Encode(dst, src)
	return dst
}
