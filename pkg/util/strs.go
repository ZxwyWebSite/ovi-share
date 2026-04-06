package util

// 字符串的两种表示
type str interface{ string | []byte }

// 拼接字符串
func Concat[T str](s ...T) string {
	return BytesToString(ConcatB(s...))
}

// 拼接字符串
func ConcatB[T str](s ...T) []byte {
	b := MakeNoZero(LenLoop(s))
	var p int
	for i, r := 0, len(s); i < r; i++ {
		p += copy(b[p:], s[i])
	}
	return b
}
